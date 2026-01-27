package helpers

// HelperRef points to a helper implementation.
type HelperRef struct {
	ImportPath string
	Ident      string
}

// Registry returns a map of helper names to their HelperRef for use in compiler.Options.
// This provides all Handlebars core + handlebars-helpers 7.4 helpers.
// All helpers are in the github.com/andriyg76/go-hbars/helpers/handlebars package.
func Registry() map[string]HelperRef {
	importPath := "github.com/andriyg76/go-hbars/helpers/handlebars"
	runtimePath := "github.com/andriyg76/go-hbars/runtime"
	return map[string]HelperRef{
		// Layout block helpers (partial/block) — use runtime.Blocks / ctx.Output
		"block":   {ImportPath: runtimePath, Ident: "Block"},
		"partial": {ImportPath: runtimePath, Ident: "Partial"},
		// String helpers
		"upper":        {ImportPath: importPath, Ident: "Upper"},
		"lower":        {ImportPath: importPath, Ident: "Lower"},
		"capitalize":   {ImportPath: importPath, Ident: "Capitalize"},
		"capitalizeAll": {ImportPath: importPath, Ident: "CapitalizeAll"},
		"truncate":     {ImportPath: importPath, Ident: "Truncate"},
		"reverse":      {ImportPath: importPath, Ident: "Reverse"},
		"replace":      {ImportPath: importPath, Ident: "Replace"},
		"stripTags":    {ImportPath: importPath, Ident: "StripTags"},
		"stripQuotes":  {ImportPath: importPath, Ident: "StripQuotes"},
		"join":         {ImportPath: importPath, Ident: "Join"},
		"split":        {ImportPath: importPath, Ident: "Split"},
		
		// Comparison helpers
		"eq":  {ImportPath: importPath, Ident: "Eq"},
		"ne":  {ImportPath: importPath, Ident: "Ne"},
		"lt":  {ImportPath: importPath, Ident: "Lt"},
		"lte": {ImportPath: importPath, Ident: "Lte"},
		"gt":  {ImportPath: importPath, Ident: "Gt"},
		"gte": {ImportPath: importPath, Ident: "Gte"},
		"and": {ImportPath: importPath, Ident: "And"},
		"or":  {ImportPath: importPath, Ident: "Or"},
		"not": {ImportPath: importPath, Ident: "Not"},
		
		// Date helpers
		"formatDate": {ImportPath: importPath, Ident: "FormatDate"},
		"now":        {ImportPath: importPath, Ident: "Now"},
		"ago":        {ImportPath: importPath, Ident: "Ago"},
		
		// Collection helpers
		"lookup":  {ImportPath: importPath, Ident: "Lookup"},
		"default": {ImportPath: importPath, Ident: "Default"},
		"length":  {ImportPath: importPath, Ident: "Length"},
		"first":   {ImportPath: importPath, Ident: "First"},
		"last":    {ImportPath: importPath, Ident: "Last"},
		"inArray": {ImportPath: importPath, Ident: "InArray"},
		
		// Math helpers
		"add":      {ImportPath: importPath, Ident: "Add"},
		"subtract": {ImportPath: importPath, Ident: "Subtract"},
		"multiply": {ImportPath: importPath, Ident: "Multiply"},
		"divide":   {ImportPath: importPath, Ident: "Divide"},
		"modulo":   {ImportPath: importPath, Ident: "Modulo"},
		"floor":    {ImportPath: importPath, Ident: "Floor"},
		"ceil":     {ImportPath: importPath, Ident: "Ceil"},
		"round":    {ImportPath: importPath, Ident: "Round"},
		"abs":      {ImportPath: importPath, Ident: "Abs"},
		"min":      {ImportPath: importPath, Ident: "Min"},
		"max":      {ImportPath: importPath, Ident: "Max"},
		
		// Number helpers
		"formatNumber": {ImportPath: importPath, Ident: "FormatNumber"},
		"toInt":        {ImportPath: importPath, Ident: "ToInt"},
		"toFloat":      {ImportPath: importPath, Ident: "ToFloat"},
		"random":       {ImportPath: importPath, Ident: "Random"},
		"toFixed":     {ImportPath: importPath, Ident: "ToFixed"},
		"toString":     {ImportPath: importPath, Ident: "ToString"},
		"toNumber":    {ImportPath: importPath, Ident: "ToNumber"},
		
		// Object helpers
		"has":        {ImportPath: importPath, Ident: "Has"},
		"keys":       {ImportPath: importPath, Ident: "Keys"},
		"values":     {ImportPath: importPath, Ident: "Values"},
		"size":       {ImportPath: importPath, Ident: "Size"},
		"isEmpty":    {ImportPath: importPath, Ident: "IsEmpty"},
		"isNotEmpty": {ImportPath: importPath, Ident: "IsNotEmpty"},
		
		// URL helpers
		"encodeURI":      {ImportPath: importPath, Ident: "EncodeURI"},
		"decodeURI":      {ImportPath: importPath, Ident: "DecodeURI"},
		"stripProtocol": {ImportPath: importPath, Ident: "StripProtocol"},
		"stripQuerystring": {ImportPath: importPath, Ident: "StripQuerystring"},
	}
}

