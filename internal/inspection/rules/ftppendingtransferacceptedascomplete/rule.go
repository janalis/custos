// Package ftppendingtransferacceptedascomplete implements FtpPendingTransferAcceptedAsComplete.
package ftppendingtransferacceptedascomplete

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Wait for FTP_FINISHED before reporting success."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FtpPendingTransferAcceptedAsComplete" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFtpPendingTransferAcceptedAsComplete(ctx, n, message)
}
