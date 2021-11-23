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

type GoldQuery interface {
	// Field map key is target column, value is support operation slice
	Field() map[string][]Filter
}

type Golf struct {
	db *gorm.DB
	//ctx     context.Context
	builted bool
	Error   error
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
func (g *Golf) Build(model GoldQuery, query map[string][]string) *Golf {
	if g.db == nil {
		g.Error = errors.New("golf db is nil")
		return g
	}
	if reflect.ValueOf(model).Kind() != reflect.Ptr {
		g.Error = errors.New("model need a struct pointer")
		return g
	}
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
	queryMap, err := g.checkAndBuildQuery(lowerQuery, query)
	if err != nil {
		g.Error = err
		return g
	}
	operations, err := g.parseOperation(queryMap, goStructMap)
	if err != nil {
		g.Error = errors.Wrap(err, "parse operation error")
	}
	for _, operation := range operations {
		switch operation.Filter {
		case In, NotIn:
			fmt.Printf("%s %s (?)", operation.Column, getSQLOperation(operation.Filter))
			g.db = g.db.Where(fmt.Sprintf("%s %s (?)", operation.Column, getSQLOperation(operation.Filter)), operation.Value)
		default:
			fmt.Printf("%s %s (?)", operation.Column, getSQLOperation(operation.Filter))

			g.db = g.db.Where(fmt.Sprintf("%s %s ?", operation.Column, getSQLOperation(operation.Filter)), operation.Value)
		}
	}
	g.builted = true
	return g
}

func (g *Golf) Find(dest interface{}, conds ...interface{}) *Golf {
	if !g.builted {
		g.Error = errors.New("before call do you should call build first")
	}
	g.Error = g.db.Find(dest, conds...).Error
	return g
}

func (g *Golf) First(dest interface{}, conds ...interface{}) *Golf {
	if !g.builted {
		g.Error = errors.New("before call do you should call build first")
	}
	g.Error = g.db.First(dest, conds...).Error
	return g
}

func (g *Golf) parseOperation(queryMap map[string]ValueOperation, structMap map[string]reflect.StructField) ([]OperationWithType, error) {
	var ret []OperationWithType
	for k, v := range queryMap {
		goStruct := structMap[k]
		fmt.Printf(goStruct.Type.String())
		switch goStruct.Type.String() {
		case "int":
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

		// TODO parse Type to interface
		ret = append(ret, oper)
	}
	return ret, nil
}

// checkAndBuildQuery check url query and build  column value map
// urlQuery eg: eq_id=1
func (g *Golf) checkAndBuildQuery(lowerQuery map[string][]Filter, urlQuery map[string][]string) (map[string]ValueOperation, error) {
	var ret = make(map[string]ValueOperation)
	for k, v := range urlQuery {
		if len(strings.Split(k, querySep)) < 1 {
			return nil, fmt.Errorf("format query param failed,query param should like `eq_id=1`")
		}
		splitQuery := strings.Split(k, querySep)
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
		singleQ := ValueOperation{
			Value:  v,
			Column: queryColumn,
			Filter: filter,
		}
		ret[queryColumn] = singleQ
	}
	return ret, nil
}
