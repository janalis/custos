<?php
function binding_column($rows) { return array_column($rows, 'v', 'id'); }
function binding_combine($keys, $values) { return array_combine($keys, $values); }
function binding_replace($left, $right) { return array_replace($left, $right); }
function binding_local($row) { ['label' => $value, 'unused' => $unused] = $row; return $value; }
function binding_prepare($rows) {
    $out = [];
    foreach ($rows as $row) {
        $value = $row['label'];
        $out[] = $value;
    }
    return $out;
}
function binding_output(): void {
    echo '<p>' . binding_column([['id' => 'x', 'v' => $_GET['unused']], ['id' => 'x', 'v' => 'fixed']])['x'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">binding_column([['id' => 'x', 'v' => $_GET['label']]])['x']</error>;
    echo '<p>' . binding_combine(['x', 'x'], [$_GET['unused'], 'fixed'])['x'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">binding_combine(['x'], [$_GET['label']])['x']</error>;
    echo '<p>' . binding_replace(['x' => $_GET['unused']], ['x' => 'fixed'])['x'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">binding_replace(['x' => 'fixed'], ['x' => $_GET['label']])['x']</error>;
    echo '<p>' . binding_local(['label' => 'fixed', 'unused' => $_GET['unused']]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">binding_local(['label' => $_GET['label'], 'unused' => 'fixed'])</error>;
    echo '<p>' . binding_prepare([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">binding_prepare([['label' => $_GET['label']]])[0]</error>;
}
