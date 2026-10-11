<?php

function lateCookie(): void
{
    echo 'started';
    flush();
    <warning descr="Set response headers before committing output.">setcookie('seen', '1')</warning>;
}

function lateStatus(int $status): void
{
    echo 'started';
    flush();
    <warning descr="Set response headers before committing output.">http_response_code(503)</warning>;
    <warning descr="Set response headers before committing output.">http_response_code(response_code: $status)</warning>;
}

function wrongBranch(): void
{
    echo 'started';
    flush();
    if (headers_sent()) {
        <warning descr="Set response headers before committing output.">header('X-State: late')</warning>;
    }
    headers_sent() && <warning descr="Set response headers before committing output.">header('X-State: later')</warning>;
}

function committedAfterTheTest(): void
{
    if (!headers_sent()) {
        echo 'started';
        flush();
        <warning descr="Set response headers before committing output.">header('X-State: stale')</warning>;
    }
}
