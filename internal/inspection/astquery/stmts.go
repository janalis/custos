package astquery

import (
	"custos/internal/php/syntax"
)

// StmtList returns the statement list that directly contains s (a block,
// switch case, namespace body or the file's top level) and s's index in it;
// ok is false when s is not part of a statement list (e.g. an unbraced body).
func StmtList(f *syntax.File, s syntax.Stmt) (list []syntax.Stmt, idx int, ok bool) {
	switch p := s.Parent().(type) {
	case *syntax.File:
		if p != f {
			return nil, -1, false
		}
		list = p.Stmts
	case *syntax.Block:
		list = p.Stmts
	case *syntax.Case:
		list = p.Stmts
	case *syntax.Namespace:
		list = p.Stmts
	default:
		return nil, -1, false
	}
	// A parsed statement whose parent owns a list is one of its entries.
	return list, syntax.StmtIndex(list, s), true
}

// PrevStmt returns the statement preceding s in its statement list.
func PrevStmt(f *syntax.File, s syntax.Stmt) (syntax.Stmt, bool) {
	list, i, ok := StmtList(f, s)
	if !ok || i == 0 {
		return nil, false
	}
	return list[i-1], true
}

// NextStmt returns the statement following s in its statement list.
func NextStmt(f *syntax.File, s syntax.Stmt) (syntax.Stmt, bool) {
	list, i, ok := StmtList(f, s)
	if !ok || i+1 >= len(list) {
		return nil, false
	}
	return list[i+1], true
}

// BlockStmts returns the statements of a braced or alternative-syntax block,
// ok is false when body is not a *syntax.Block.
func BlockStmts(body syntax.Stmt) ([]syntax.Stmt, bool) {
	b, ok := body.(*syntax.Block)
	if !ok {
		return nil, false
	}
	return b.Stmts, true
}
