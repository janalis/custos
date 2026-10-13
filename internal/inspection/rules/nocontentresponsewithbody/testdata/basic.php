<?php
http_response_code(204); <warning descr="Omit response content for status 204.">echo '{"done":true}';</warning>

function deleted(string $type): void
{
    http_response_code(204);
    header('X-Request-Id: 42');
    header('Content-Type: ' . $type);
    header('HTTP/1.1 204 No Content');
    header('X-Reason: removed', true, 204);
    <warning descr="Omit response content for status 204.">echo 'removed';</warning>
}

function afterDiscardedBuffer(): void
{
    http_response_code(204);
    ob_start();
    echo 'debug';
    ob_end_clean();
    <warning descr="Omit response content for status 204.">echo 'removed';</warning>
}

function headerAfterBody(): void
{
    http_response_code(204);
    <warning descr="Omit response content for status 204.">echo 'removed';</warning>
    header('HTTP/1.1 200 OK');
}
