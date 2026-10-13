<?php
// File-level variables are globals: any function may rewrite them.
$fieldGlue = '';
$rowGlue = '';
function chooseGlue() {
    global $fieldGlue;
    $fieldGlue = ',';
    $GLOBALS['rowGlue'] = ';';
}
chooseGlue();
$fields = explode($fieldGlue, $line);
$rows = explode($rowGlue, $text);
