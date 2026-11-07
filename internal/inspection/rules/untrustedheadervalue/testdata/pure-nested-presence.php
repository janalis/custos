<?php
function nested_header_label($row) {
    return ($row['a'] ?? ['label' => 'fallback'])['label'];
}
function nested_header_output(): void {
    header('X-Value: ' . nested_header_label([]));
    header('X-Value: ' . nested_header_label(['a' => ['label' => 'fixed'], 'unused' => $_GET['unused']]));
    <error descr="Reject line breaks and validate this header value.">header('X-Value: ' . nested_header_label(['a' => ['label' => $_GET['label']]]))</error>;
}
