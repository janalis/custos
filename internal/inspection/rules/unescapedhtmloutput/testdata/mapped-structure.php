<?php
function mappedStructure($value) { return ['safe' => 'text', 'unsafe' => $value]; }
function mappedStructureCaller() {
    $rows = array_map('mappedStructure', [$_GET['value']]);
    echo '<p>' . $rows[0]['safe'] . '</p>';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$rows[0]['unsafe']</error> . '</p>';
    $escaped = array_map(fn($value, $flags) => htmlspecialchars($value, $flags), [$_GET['value']], [ENT_QUOTES]);
    echo '<p>' . $escaped[0] . '</p>';
    $padded = array_map(fn($value, $extra) => $extra, [$_GET['value']], []);
    echo '<p>' . $padded[0] . '</p>';
}
