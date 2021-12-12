package golf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var OperationMap = map[Filter]Filter{
	Equal:   Equal,
	Lt:      Lt,
	Lte:     Lte,
	Gt:      Gt,
	Gte:     Gte,
	Like:    Like,
	NotLike: NotLike,
	In:      In,
	NotIn:   NotIn,
}

type GolfQuery interface {
	// Field map key is target column, value is support operation slice
	Field() map[string][]Filter
}

type Golf struct {
	db            *gorm.DB
	isBuild       bool
	Error         error
	filters       []OperationWithType
	originalQuery map[string][]string
	count         int64
	offset        int64
}

func NewGolf(db *gorm.DB) *Golf {
	return &Golf{
		db: db,
	}
}

func (g *Golf) GetGormDB() *gorm.DB {
	return g.db
}

// Build Golf wile call checkQuery before generate real query
func (g *Golf) Build(model GolfQuery, query map[string][]string) *Golf {
	if g.db == nil {
		g.Error = errors.New("golf db is nil")
		return g
	}
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		g.Error = errors.New("model need a struct pointer")
		return g
	}
	g.originalQuery = query
	fields := model.Field()
	elem := reflect.TypeOf(model).Elem()
	var lowerQuery = make(map[string][]Filter)
	var goStructMap = make(map[string]reflect.StructField)
	for k, v := range fields {
		localStruct, ok := elem.FieldByName(k)
		if !ok {
			g.Error = fmt.Errorf(fmt.Sprintf("invalidate model field: %s", k))
			break
		}
		lowerColumn := strcase.ToSnake(k)
		goStructMap[lowerColumn] = localStruct
		// TODO Origin SQL
		lowerQuery[lowerColumn] = v
	}
	queryMap, err := g.checkAndBuildQuery(lowerQuery)
	if err != nil {
		g.Error = err
		return g
	}
	g.filters, g.Error = g.parseOperation(queryMap, goStructMap)
	return g.buildGormQuery()
}

func (g *Golf) buildGormQuery() *Golf {
	if err := g.buildPagination().Error; err != nil {
		return g
	}
	for _, operation := range g.filters {
		switch operation.Filter {
		case In, NotIn:
			g.db = g.db.Where(fmt.Sprintf("%s %s (?)", operation.Column, getSQLOperation(operation.Filter)), operation.Value)
		default:
			g.db = g.db.Where(fmt.Sprintf("%s %s ?", operation.Column, getSQLOperation(operation.Filter)), operation.Value)
		}
	}
	g.isBuild = true
	return g
}

func (g *Golf) Find(dest interface{}, conds ...interface{}) *Golf {

	if g.Error != nil {
		return g
	}

	if !g.isBuild {
		g.Error = errors.New("before call find, you should call build first")
	}
	if g.offset != 0 || g.count != 0 {
		g.Error = g.db.Limit(int(g.count)).Offset(int(g.offset)).Find(dest, conds...).Error
		return g
	}
	g.Error = g.db.Find(dest, conds...).Error
	return g
}

func (g *Golf) First(dest interface{}, conds ...interface{}) *Golf {
	if g.Error != nil {
		return g
	}
	if !g.isBuild {
		g.Error = errors.New("before call first, you should call build first")
	}

	g.Error = g.db.First(dest, conds...).Error
	return g
}

func (g *Golf) parseOperation(queryMap []ValueOperation, structMap map[string]reflect.StructField) ([]OperationWithType, error) {
	var ret []OperationWithType
	for _, v := range queryMap {
		goStruct := structMap[v.Column]
		switch goStruct.Type.String() {
		case "int", "int64", "int32", "uint", "uint64":
			i, err := strconv.ParseInt(v.Value.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			v.Value = i
		case "string":
			v.Value = v.Value.(string)
		}
		oper := OperationWithType{
			ValueOperation: v,
		}

		ret = append(ret, oper)
	}
	return ret, nil
}

// checkAndBuildQuery check url query and build  column value map
// urlQuery eg: eq_id=1
func (g *Golf) checkAndBuildQuery(lowerQuery map[string][]Filter) ([]ValueOperation, error) {
	var ret []ValueOperation
	for k, v := range g.originalQuery {
		if len(strings.Split(k, querySep)) < 1 {
			return nil, fmt.Errorf("format query param failed,query param should like `eq_id=1`")
		}
		splitQuery := strings.Split(k, querySep)
		if len(splitQuery) <= 1 {
			continue
		}
		filter, ok := OperationMap[Filter(splitQuery[0])]
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("un support oper: %s", splitQuery[1]))
		}
		// extract real query column support for gte_user_id=1
		queryColumn := strings.Replace(k, fmt.Sprintf("%s%s", splitQuery[0], querySep), "", 1)
		supportFilter, ok := lowerQuery[queryColumn]
		if !ok {
			return nil, fmt.Errorf(fmt.Sprintf("Undefined field %s", queryColumn))
		}
		var support bool
		for _, q := range supportFilter {
			if q == filter {
				support = true
			}
		}
		if !support {
			return nil, fmt.Errorf(fmt.Sprintf("field:%s un support operation: %s", splitQuery[1], splitQuery[0]))
		}
		for _, vv := range v {
			singleQ := ValueOperation{
				Value:  vv,
				Column: queryColumn,
				Filter: filter,
			}
			ret = append(ret, singleQ)
		}

	}
	return ret, nil
}

func (g *Golf) buildPagination() *Golf {
	for k, v := range g.originalQuery {
		if len(v) > 0 {
			switch k {
			case "offset":
				offset, err := strconv.ParseInt(v[0], 10, 64)
				if err != nil {
					g.Error = err
				}
				g.offset = offset
			case "count":
				offset, err := strconv.ParseInt(v[0], 10, 64)
				if err != nil {
					g.Error = err
				}
				g.count = offset
			default:
				g.count = 10
				g.offset = 0
			}
		}

	}
	return g
}
