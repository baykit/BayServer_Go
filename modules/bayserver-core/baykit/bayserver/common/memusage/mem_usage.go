package memusage

/****************************************/
/*  Type MemUsage                       */
/****************************************/

type MemUsage interface {
	PrintUsage(indent int)
}

/****************************************/
/*  Static methods                      */
/****************************************/

var NewMemUsage func(agtId int) MemUsage
var Get func(agtId int) MemUsage
