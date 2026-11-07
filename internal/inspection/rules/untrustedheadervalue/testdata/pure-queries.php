<?php
function header_quiet_label($record) {
    return $record['label'] ?? 'fallback';
}
function header_pure_queries(): void {
    header('X-Value: ' . header_quiet_label([]));
    header('X-Value: ' . header_quiet_label(['label' => null]));
    header('X-Value: ' . header_quiet_label(['label' => 'fixed', 'unused' => $_GET['unused']]));
    <error descr="Reject line breaks and validate this header value.">header('X-Value: ' . header_quiet_label(['label' => $_GET['label']]))</error>;
}
