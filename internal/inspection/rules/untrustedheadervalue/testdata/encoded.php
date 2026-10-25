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
