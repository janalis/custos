<?php
function filterHeaderWrapper($value) {
    array_filter([$value], fn($entry) => header($entry));
}
<error descr="Reject line breaks and validate this header value.">filterHeaderWrapper($_GET['header'])</error>;
<error descr="Reject line breaks and validate this header value.">array_filter([$_GET['header']], 'header')</error>;
array_filter(['X-Fixed' => $_GET['header']], fn($key) => header($key), ARRAY_FILTER_USE_KEY);
