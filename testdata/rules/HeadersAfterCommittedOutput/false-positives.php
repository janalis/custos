<?php

function readStatus(): int
{
    echo 'started';
    flush();
    $status = http_response_code();
    $same = http_response_code(0);

    return $status ?: $same;
}

function guardedBlock(string $chunk): void
{
    echo $chunk;
    flush();
    if (!headers_sent()) {
        header('X-State: ready');
    }
    if (headers_sent() === false) {
        setcookie('seen', '1');
    }
}

function guardedElse(string $chunk): void
{
    echo $chunk;
    flush();
    if (headers_sent()) {
        $late = true;
    } else {
        header('X-State: ready');
    }
}

function guardedExpression(string $chunk): void
{
    echo $chunk;
    flush();
    headers_sent() or header('X-State: ready');
    !headers_sent() && http_response_code(202);
    $done = headers_sent() ? null : header('X-Mode: stream');
}

function guardedReturn(string $chunk): void
{
    echo $chunk;
    flush();
    if (headers_sent()) {
        return;
    }
    header('X-State: ready');
}

function buffered(): void
{
    ob_start();
    echo 'started';
    flush();
    header('X-State: ready');
}
