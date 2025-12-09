package work

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (t *Work) Sql(ctx *Context) error {
	cfg := SqlQuery{ctx: ctx}

	switch t.Kind {
	case KindValue:
		// xử lý giá trị đơn giản nếu cần

	case KindFull:
		for k, v := range t.Config {
			switch k {
			case "count":
				cfg.Type = "count"
				var field string = "*"
				switch val := v.(type) {
				case string:
					rendered, _ := ctx.render(val)
					if rendered != "" && rendered != "all" && rendered != "*" {
						field = rendered
					}
				case []string:
					if len(val) > 0 {
						rendered, _ := ctx.render(val[0])
						field = rendered
					}
				case []interface{}:
					if len(val) > 0 {
						s, _ := ctx.render(ToString(val[0]))
						field = s
					}
				}
				cfg.Fields = append(cfg.Fields, field)
			case "select":
				cfg.Type = "select"
				switch val := v.(type) {
				case string:
					rendered, _ := ctx.render(val)
					if rendered == "all" || rendered == "*" {
						cfg.Fields = append(cfg.Fields, "*")
					} else {
						cfg.Fields = append(cfg.Fields, rendered)
					}
				case []string:
					for _, s := range val {
						rendered, _ := ctx.render(s)
						cfg.Fields = append(cfg.Fields, rendered)
					}
				case []interface{}:
					for _, item := range val {
						s, _ := ctx.render(ToString(item))
						cfg.Fields = append(cfg.Fields, s)
					}
				}

			case "from", "table":
				cfg.Table, _ = ctx.render(ToString(v))

			case "as":
				cfg.As, _ = ctx.render(ToString(v))

			case "order":
				ob, err := NewOrderBy(v)
				if err != nil {
					return err
				}
				cfg.Order = ob

			case "limit":
				if n, ok := v.(int); ok {
					cfg.Limit = n
				} else if s, ok := v.(string); ok {
					rendered, _ := ctx.render(s)
					if val, err := strconv.Atoi(rendered); err == nil {
						cfg.Limit = val
					}
				}

			case "offset":
				if n, ok := v.(int); ok {
					cfg.Offset = n
				} else if s, ok := v.(string); ok {
					rendered, _ := ctx.render(s)
					if val, err := strconv.Atoi(rendered); err == nil {
						cfg.Offset = val
					}
				}

			case "where":
				cn, err := parseWhereNode(v, ctx)
				if err != nil {
					return err
				}
				cfg.Where = &cn

			case "insert", "update":
				cfg.Type = ToString(k)
				cfg.Table = ToString(v)
			case "values":
				if m, ok := v.(map[string]interface{}); ok {
					cfg.Values = map[string]interface{}{}
					for key, val := range m {
						rendered, _ := ctx.render(ToString(val))
						cfg.Values[key] = rendered
					}
				} else {
					return fmt.Errorf("values must be map[string]interface{} for %s", k)
				}

			case "returning":

				switch val := v.(type) {
				case string:
					rendered, _ := ctx.render(val)
					if rendered == "all" || rendered == "*" {
						cfg.Returning = append(cfg.Returning, "*")
					} else {
						cfg.Returning = append(cfg.Returning, rendered)
					}
				case []string:
					for _, s := range val {
						rendered, _ := ctx.render(s)
						cfg.Returning = append(cfg.Returning, rendered)
					}
				case []interface{}:
					for _, item := range val {
						s, _ := ctx.render(ToString(item))
						cfg.Returning = append(cfg.Returning, s)
					}
				}

			case "delete":
				cfg.Type = "delete"

			default:
				return fmt.Errorf("unknown key in config: %s", k)
			}
		}

	default:
		return fmt.Errorf("type is not a request/fetch type: %s", t.Type)
	}

	result, err := cfg.Execute()
	if err != nil {
		return err
	}

	return ctx.As(cfg.As, result)
}

// SqlQuery đại diện cho một truy vấn động
type SqlQuery struct {
	ctx  *Context
	Name string
	As   string

	DB   string // tên database / adapter
	Type string // "select", "insert", "update", "delete"

	Alias  string
	Fields []string

	Joins  []Join
	Where  *ConditionNode
	Group  []string
	Having *ConditionNode
	Order  []OrderBy

	Returning []string
	Table     string
	Values    map[string]interface{} // insert/update

	Limit  int
	Offset int
}

// Join
type Join struct {
	Type  string // "inner", "left", "right"
	Table string
	On    string
}

// OrderBy
type OrderBy struct {
	Field     string
	Direction string // ASC / DESC
}

func NewOrderBy(input interface{}) ([]OrderBy, error) {
	var orders []OrderBy

	switch val := input.(type) {
	case string:
		// "age DESC" hoặc "name ASC"
		parts := strings.Fields(val)
		if len(parts) == 1 {
			orders = append(orders, OrderBy{Field: parts[0], Direction: "ASC"})
		} else if len(parts) >= 2 {
			orders = append(orders, OrderBy{Field: parts[0], Direction: strings.ToUpper(parts[1])})
		}

	case []string:
		for _, s := range val {
			o, err := NewOrderBy(s)
			if err != nil {
				return nil, err
			}
			orders = append(orders, o...)
		}

	case []interface{}:
		for _, item := range val {
			o, err := NewOrderBy(item)
			if err != nil {
				return nil, err
			}
			orders = append(orders, o...)
		}

	case map[string]interface{}:
		// {"field":"age", "direction":"desc"}
		field := ToString(val["field"])
		dir := strings.ToUpper(ToString(val["direction"]))
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		orders = append(orders, OrderBy{Field: field, Direction: dir})

	case []map[string]interface{}:
		for _, m := range val {
			o, err := NewOrderBy(m)
			if err != nil {
				return nil, err
			}
			orders = append(orders, o...)
		}

	default:
		return nil, fmt.Errorf("unsupported order type: %T", val)
	}

	return orders, nil
}

// ConditionNode hỗ trợ AND/OR lồng nhau
type ConditionNode struct {
	And  []ConditionNode
	Or   []ConditionNode
	Leaf *LeafCondition
}

type LeafCondition struct {
	Field    string
	Operator string
	Value    interface{}
}

// buildCondition sinh ra SQL string và args từ ConditionNode
func buildCondition(node ConditionNode) (string, []interface{}) {
	if node.Leaf != nil {
		return fmt.Sprintf("%s %s ?", node.Leaf.Field, node.Leaf.Operator), []interface{}{node.Leaf.Value}
	}

	parts := []string{}
	args := []interface{}{}

	for _, andNode := range node.And {
		sql, a := buildCondition(andNode)
		parts = append(parts, fmt.Sprintf("(%s)", sql))
		args = append(args, a...)
	}

	if len(node.Or) > 0 {
		orParts := []string{}
		for _, orNode := range node.Or {
			sql, a := buildCondition(orNode)
			orParts = append(orParts, fmt.Sprintf("(%s)", sql))
			args = append(args, a...)
		}
		parts = append(parts, "("+strings.Join(orParts, " OR ")+")")
	}

	return strings.Join(parts, " AND "), args
}

// ToGorm build GORM query từ SqlQuery
func (q *SqlQuery) Database() *gorm.DB {
	db := q.ctx.database.Get()

	if db == nil {
		return nil
	}

	table := q.Table
	if q.Alias != "" {
		table = fmt.Sprintf("%s AS %s", q.Table, q.Alias)
	}
	g := db.Table(table)

	if len(q.Fields) > 0 && strings.ToLower(q.Type) == "select" {
		g = g.Select(strings.Join(q.Fields, ", "))
	}

	// Joins
	for _, j := range q.Joins {
		g = g.Joins(fmt.Sprintf("%s JOIN %s ON %s", strings.ToUpper(j.Type), j.Table, j.On))
	}

	// Where
	if q.Where != nil {
		sql, args := buildCondition(*q.Where)
		g = g.Where(sql, args...)
	}

	// Group / Having
	if len(q.Group) > 0 {
		g = g.Group(strings.Join(q.Group, ", "))
	}
	if q.Having != nil {
		sql, args := buildCondition(*q.Having)
		g = g.Having(sql, args...)
	}

	// Order
	for _, o := range q.Order {
		g = g.Order(fmt.Sprintf("%s %s", o.Field, o.Direction))
	}

	// Limit / Offset
	if q.Limit > 0 {
		g = g.Limit(q.Limit)
	}
	if q.Offset > 0 {
		g = g.Offset(q.Offset)
	}

	return g
}

func (q *SqlQuery) Execute() (interface{}, error) {

	database := q.Database()

	if database == nil {
		return nil, fmt.Errorf("SqlQuery.db is nil")
	}

	switch strings.ToLower(q.Type) {
	case "select":
		var results []map[string]interface{}
		if err := database.Find(&results).Error; err != nil {
			return nil, err
		}
		return results, nil

	case "count":
		var total int64
		if err := database.Count(&total).Error; err != nil {
			return nil, err
		}
		return total, nil
	case "insert":
		if q.Table == "" {
			return nil, fmt.Errorf("missing table")
		}

		var retCols []clause.Column

		// --- CASE 3: Không truyền Returning -> trả về row affected ---
		if len(q.Returning) == 0 {
			query := database.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Table(q.Table).Create(q.Values)
			})

			result := database.Exec(query)

			return map[string]interface{}{
				"rows_affected": result.RowsAffected,
			}, result.Error
		}

		// --- CASE 1 & 2: Có truyền Returning ---

		// CASE 1: Returning = ["*"]
		if len(q.Returning) == 1 && q.Returning[0] == "*" {
			retCols = nil // RETURNING *
		} else {
			// CASE 2: Returning theo từng cột
			retCols = make([]clause.Column, 0, len(q.Returning))
			for _, col := range q.Returning {
				retCols = append(retCols, clause.Column{Name: col})
			}
		}

		// Build SQL
		query := database.ToSQL(func(tx *gorm.DB) *gorm.DB {
			if retCols == nil {
				return tx.Table(q.Table).
					Clauses(clause.Returning{}). // RETURNING *
					Create(q.Values)
			}

			return tx.Table(q.Table).
				Clauses(clause.Returning{Columns: retCols}).
				Create(q.Values)
		})

		var out map[string]interface{}
		err := database.Raw(query).Scan(&out).Error
		return out, err

	case "update":
		if err := database.Updates(q.Values).Error; err != nil {
			return nil, err
		}
		return nil, nil

	case "delete":
		if err := database.Delete(nil).Error; err != nil {
			return nil, err
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown query type: %s", q.Type)
	}
}

// ToString chuyển interface{} sang string một cách an toàn
func ToString(val interface{}) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		// fallback: dùng fmt.Sprintf
		return fmt.Sprintf("%v", v)
	}
}

// ToStringSlice chuyển []interface{} sang []string
func ToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = ToString(v) // dùng hàm ToString trước đó
	}
	return result
}

// helper map operator từ map key
func mapOperator(op string) string {
	switch strings.ToLower(op) {
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "lt":
		return "<"
	case "lte":
		return "<="
	case "neq", "ne":
		return "!="
	case "eq":
		return "="
	case "like":
		return "LIKE"
	default:
		return op // mặc định giữ nguyên
	}
}

func parseLeafCondition(field string, val interface{}, ctx *Context) ([]ConditionNode, error) {
	var nodes []ConditionNode

	// Render template nếu là string
	var v interface{} = val
	if s, ok := val.(string); ok {
		rendered, err := ctx.render(s)
		if err != nil {
			return nil, fmt.Errorf("render field %s: %w", field, err)
		}
		v = rendered
	}

	switch val := v.(type) {
	case string:
		parts := strings.Fields(val)
		if len(parts) == 1 {
			nodes = append(nodes, ConditionNode{
				Leaf: &LeafCondition{Field: field, Operator: "=", Value: val},
			})
		} else if len(parts) >= 2 {
			op := parts[0]
			value := strings.Join(parts[1:], " ")
			nodes = append(nodes, ConditionNode{
				Leaf: &LeafCondition{Field: field, Operator: op, Value: value},
			})
		}
	case []interface{}:
		nodes = append(nodes, ConditionNode{
			Leaf: &LeafCondition{Field: field, Operator: "IN", Value: val},
		})
	case map[string]interface{}:
		for k, val := range val {
			op := mapOperator(k)
			nodes = append(nodes, ConditionNode{
				Leaf: &LeafCondition{Field: field, Operator: op, Value: val},
			})
		}
	default:
		nodes = append(nodes, ConditionNode{
			Leaf: &LeafCondition{Field: field, Operator: "=", Value: val},
		})
	}

	return nodes, nil
}

func parseWhereNode(input interface{}, ctx *Context) (ConditionNode, error) {
	node := ConditionNode{}

	switch v := input.(type) {
	case map[string]interface{}:
		for k, val := range v {
			lowerKey := strings.ToLower(k)
			if lowerKey == "and" || lowerKey == "or" {
				arr, ok := val.([]interface{})
				if !ok {
					return node, fmt.Errorf("%s value must be array", k)
				}
				childNodes := []ConditionNode{}
				for _, item := range arr {
					cn, err := parseWhereNode(item, ctx)
					if err != nil {
						return node, err
					}
					childNodes = append(childNodes, cn)
				}
				if lowerKey == "and" {
					node.And = append(node.And, childNodes...)
				} else {
					node.Or = append(node.Or, childNodes...)
				}
			} else {
				leafNodes, err := parseLeafCondition(k, val, ctx)
				if err != nil {
					return node, err
				}
				node.And = append(node.And, leafNodes...)
			}
		}
	default:
		return node, fmt.Errorf("unsupported where type: %T", input)
	}

	return node, nil
}
