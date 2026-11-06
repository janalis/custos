<?php
function composition_join($pieces) { return implode(',', $pieces); }
function composition_split($text) { return explode(',', $text); }
function composition_column($rows) { return array_column($rows, 'label'); }
function composition_merge($left, $right) { return array_merge($left, $right); }
function composition_closure($rows) {
    return array_map(function ($row) { return $row['label']; }, $rows);
}
function composition_find($rows) {
    foreach ($rows as ['keep' => $keep, 'label' => $label]) {
        if ($keep) { return $label; }
    }
    return 'fixed';
}
function composition_output(): void {
    echo '<p>' . composition_join(['first', 'second']);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_join([$_GET['label']])</error>;
    echo '<p>' . composition_split('first,second')[1];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_split($_GET['label'])[0]</error>;
    echo '<p>' . composition_column([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_column([['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . composition_merge(['label' => $_GET['unused']], ['label' => 'fixed'])['label'];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_merge(['label' => 'fixed'], ['label' => $_GET['label']])['label']</error>;
    echo '<p>' . composition_closure([['label' => 'fixed', 'unused' => $_GET['unused']]])[0];
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_closure([['label' => $_GET['label']]])[0]</error>;
    echo '<p>' . composition_find([['keep' => true, 'label' => 'fixed'], []]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">composition_find([['keep' => true, 'label' => $_GET['label']]])</error>;
}
