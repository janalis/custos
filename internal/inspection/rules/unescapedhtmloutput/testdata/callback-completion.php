<?php
function rowFactory() { return fn($value) => ['safe' => 'text', 'unsafe' => $value]; }
function callbackCompletionCaller() {
    $callback = rowFactory();
    $row = $callback($_GET['value']);
    echo '<p>' . $row['safe'] . '</p>';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$row['unsafe']</error> . '</p>';
    $mapped = array_map(...['htmlspecialchars', [$_GET['value']]]);
    echo '<p>' . $mapped[0] . '</p>';
    $zipped = array_map(null, ['text'], [$_GET['value']]);
    echo '<p>' . $zipped[0][0] . '</p>';
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$zipped[0][1]</error> . '</p>';
    $filtered = array_filter([$_GET['value']], fn($value) => htmlspecialchars($value));
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$filtered[0]</error> . '</p>';
    $coerced = array_filter([$_GET['value']], fn($value): string => -0.0);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">$coerced[0]</error> . '</p>';
}
