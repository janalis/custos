<?php

function statusLine(): void
{
    http_response_code(204);
    header('HTTP/1.1 200 OK');
    echo 'body';
}

function explicitStatus(): void
{
    http_response_code(204);
    header('X-Reason: fallback', true, 500);
    echo 'body';
}

function redirect(string $target): void
{
    http_response_code(204);
    header('Location: ' . $target);
    echo 'Redirecting';
}

function statusField(): void
{
    http_response_code(204);
    header('Status: 200');
    echo 'body';
}

function unreadableHeader(string $line): void
{
    http_response_code(204);
    header($line);
    echo 'body';
}

function conditionalStatus(bool $failed): void
{
    http_response_code(204);
    if ($failed) {
        header('HTTP/1.1 500 Internal Server Error');
    }
    echo 'body';
}

function repeated(array $rows): void
{
    http_response_code(204);
    foreach ($rows as $row) {
        echo 'row';
        header('HTTP/1.1 200 OK');
    }
}

function discarded(): void
{
    http_response_code(204);
    ob_start();
    echo 'debug';
    ob_end_clean();
}

function captured(): string
{
    http_response_code(204);
    ob_start();
    echo 'template';
    return ob_get_clean();
}

function emptyBody(): void
{
    http_response_code(204);
    echo '';
}

function otherStatus(): void
{
    http_response_code(200);
    echo 'body';
}
