<?php
namespace Report;

function implode($pieces, $glue) { return ''; }

$rows  = implode($cells, ';');
$cols  = <weak_warning descr="Pass the separator as the first argument of implode().">\IMPLODE($cells, ';')</weak_warning>;

namespace Export;

$line  = <weak_warning descr="Pass the separator as the first argument of implode().">Implode($fields, ',')</weak_warning>;
