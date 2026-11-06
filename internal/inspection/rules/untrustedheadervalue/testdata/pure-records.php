<?php
function header_record_first($values = ['safe']) {
    return $values[0];
}
function header_record_wrap($values) {
    return header_record_first(...[$values]);
}
function header_pure_records(): void {
    header('X-Value: ' . header_record_first());
    header('X-Value: ' . header_record_wrap(['safe', $_GET['unused']]));
    <error descr="Reject line breaks and validate this header value.">header('X-Value: ' . header_record_wrap([$_GET['value'], 'safe']))</error>;
}
