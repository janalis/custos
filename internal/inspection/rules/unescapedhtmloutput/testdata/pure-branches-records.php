<?php
function html_record_first($values = ['safe']) {
    return $values[0];
}
function html_record_wrap($values) {
    return html_record_first(...[$values]);
}
function html_nonblank($value) {
    if (strlen(trim($value)) === 0) {
        return false;
    }
    return true;
}
function html_pure_branches_records(): void {
    echo '<p>' . html_record_first();
    echo '<p>' . html_record_first(['safe', $_GET['unused']]);
    echo '<p>' . html_record_wrap(['safe', $_GET['unused']]);
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_record_first([$_GET['value'], 'safe'])</error>;
    echo '<p>' . <error descr="Escape untrusted text for its HTML output context.">html_record_wrap([$_GET['value'], 'safe'])</error>;
    $kept = array_filter([' ', 'visible'], 'html_nonblank');
    echo '<p>' . $kept[1];
}
