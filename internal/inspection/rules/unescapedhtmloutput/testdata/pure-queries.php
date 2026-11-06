<?php
function html_quiet_label($record) {
    return $record['label'] ?? 'fallback';
}
function html_pure_queries(): void {
    echo '<p>' . html_quiet_label([]);
    echo '<p>' . html_quiet_label(['label' => null]);
    echo '<p>' . html_quiet_label(['label' => 'fixed', 'unused' => $_GET['unused']]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_quiet_label(['label' => $_GET['label']])</error>;
}
