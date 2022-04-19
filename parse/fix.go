package parse

// 检查以下错误：
// 1. 同一个message不能有相同的成员
// 2. 同一个service不能有相同的method
// 3. 不能有相同的message(saveMessage时检查)
// 4. 不能有相同的service(saveService时检查)
// 5. 一个方法的请求参数只能有一个stream
// 6. message成员不能是stream类型

func fixSymbols(syms *Symbols) {

}
