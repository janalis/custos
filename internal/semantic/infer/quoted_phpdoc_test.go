package infer_test

import "testing"

func TestQuotedPHPDocShapeInference(t *testing.T) {
	check(t, `<?php
/** @param array{'first name': int, 'a}b': string, "x>y": bool, 'nested': array{'inner key': float}} $row */
function readRow($row) {
 t('space', $row['first name']);
 t('brace', $row['a}b']);
 t('angle', $row['x>y']);
 t('nested', $row['nested']['inner key']);
 /** @var array{'local key': string} $local */
 $local = loadRow();
 t('local', $local['local key']);
}
`, map[string]string{
		"space": "int", "brace": "string", "angle": "bool", "nested": "float", "local": "string",
	})
}

func TestQuotedPHPDocCrossFileReturn(t *testing.T) {
	checkWith(t, map[string]string{"rows.php": `<?php
namespace Data;
/** @return array{'first name': string, 'a}b': int} */
function row() { return []; }
class Rows {
 /** @return array{'display name': string, "x>y": float} */
 public function next() { return []; }
}
`}, `<?php
use Data\Rows;
$row = \Data\row();
t('functionSpace', $row['first name']);
t('functionBrace', $row['a}b']);
$rows = new Rows();
t('methodSpace', $rows->next()['display name']);
t('methodAngle', $rows->next()['x>y']);
`, map[string]string{
		"functionSpace": "string", "functionBrace": "int", "methodSpace": "string", "methodAngle": "float",
	})
}
