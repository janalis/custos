<?php
header('Location: /search?' . http_build_query(['q' => $_GET['q'], 'page' => $_GET['page']]));
header('X-Token: ' . urlencode($_GET['token']));
header('X-Raw: ' . rawurlencode($_GET['raw']));
header('X-Blob: ' . base64_encode($_POST['blob']));
header('X-Hex: ' . bin2hex($_POST['hex']));
<error descr="Reject line breaks and validate this header value.">header('X-Back: ' . urldecode($_GET['back']))</error>;
header('Location: /a?' . http_build_query($_GET, '', '&'));
header('Location: /b?' . http_build_query($_GET, arg_separator: '&amp;'));
<error descr="Reject line breaks and validate this header value.">header('Location: /c?' . http_build_query($_GET, '', $_GET['sep']))</error>;
<error descr="Reject line breaks and validate this header value.">header('Location: /d?' . http_build_query($_GET, $_GET['prefix']))</error>;
<error descr="Reject line breaks and validate this header value.">header('Location: /e?' . http_build_query($_GET, '', "\r\n"))</error>;
function redirect(array $query, string $separator, array $rest)
{
    <error descr="Reject line breaks and validate this header value.">header('Location: /h?' . http_build_query($_GET, ...$rest))</error>;
    <error descr="Reject line breaks and validate this header value.">header('Location: /g?' . http_build_query($_GET, '', $separator))</error>;
}
header('X-Data: ' . json_encode($_GET));
header('X-Data: ' . json_encode($_GET, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));
header('X-Data: ' . json_encode($_GET, flags: JSON_THROW_ON_ERROR));
<error descr="Reject line breaks and validate this header value.">header('X-Data: ' . json_encode($_GET, JSON_PRETTY_PRINT))</error>;
<error descr="Reject line breaks and validate this header value.">header('X-Data: ' . json_encode($_GET, $_GET['flags']))</error>;
