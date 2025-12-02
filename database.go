// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type SqlQuery struct {
	DB     string   // tên database / adapter, ví dụ "postgres", "mysql"
	Type   string   // "select", "update", "delete", "insert"
	Table  string   // bảng chính
	Alias  string   // optional alias
	Fields []string // select fields

	Joins  []Join                 // inner/left/right join
	Where  *ConditionNode         // AND/OR lồng nhau
	Group  []string               // group by fields
	Having *ConditionNode         // having conditions
	Order  []OrderBy              // order by fields
	Values map[string]interface{} // cho Insert/Update

	Limit  int
	Offset int
}

type ConditionNode struct {
	And  []ConditionNode // các nhánh AND
	Or   []ConditionNode // các nhánh OR
	Leaf *LeafCondition  // leaf condition thực tế
}

type LeafCondition struct {
	Field    string
	Operator string      // =, >, <, IN, LIKE, ...
	Value    interface{} // giá trị
}

type Join struct {
	Type  string // "inner", "left", "right"
	Table string // bảng join + alias
	On    string // điều kiện join
}

type OrderBy struct {
	Field     string
	Direction string // ASC / DESC
}

func applyConditionNode(db *gorm.DB, node ConditionNode) *gorm.DB {
	if node.Leaf != nil {
		return db.Where(fmt.Sprintf("%s %s ?", node.Leaf.Field, node.Leaf.Operator), node.Leaf.Value)
	}

	// AND node
	for _, andNode := range node.And {
		db = applyConditionNode(db, andNode)
	}

	// OR node
	if len(node.Or) > 0 {
		var orDB *gorm.DB
		for i, orNode := range node.Or {
			if i == 0 {
				orDB = applyConditionNode(db.Session(&gorm.Session{NewDB: true}), orNode)
			} else {
				orDB = orDB.Or(applyConditionNode(db.Session(&gorm.Session{NewDB: true}), orNode))
			}
		}
		db = db.Where(orDB)
	}

	return db
}

func BuildGormQuery(db *gorm.DB, q SqlQuery) *gorm.DB {
	g := db.Table(q.Table)
	if q.Alias != "" {
		g = g.Table(fmt.Sprintf("%s AS %s", q.Table, q.Alias))
	}

	if len(q.Fields) > 0 {
		g = g.Select(strings.Join(q.Fields, ", "))
	}

	// Joins
	for _, j := range q.Joins {
		g = g.Joins(fmt.Sprintf("%s JOIN %s ON %s", strings.ToUpper(j.Type), j.Table, j.On))
	}

	// Where
	if q.Where != nil {
		g = applyConditionNode(g, *q.Where)
	}

	// Group
	if len(q.Group) > 0 {
		g = g.Group(strings.Join(q.Group, ", "))
	}

	// Having
	if q.Having != nil {
		g = applyConditionNode(g, *q.Having)
	}

	// Order
	for _, o := range q.Order {
		g = g.Order(fmt.Sprintf("%s %s", o.Field, o.Direction))
	}

	// Limit / Offset
	if q.Limit > 0 {
		g = g.Limit(q.Limit)
	}
	g = g.Offset(q.Offset)

	return g
}

func ExecuteQuery(db *gorm.DB, q SqlQuery) ([]map[string]interface{}, error) {
	gormQuery := BuildGormQuery(db, q)

	var results []map[string]interface{}
	if err := gormQuery.Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
