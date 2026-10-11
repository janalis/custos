<?php
enum Stage: string
{
    case Draft = 'draft';
    case Live = 'live';
}

function fromCatchBlock()
{
    try {
        return load();
    } catch (ValueError $e) {
        return <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_GET['stage'])</warning>;
    }
}

function fromFinallyBlock()
{
    try {
        return load();
    } catch (ValueError $e) {
        return null;
    } finally {
        <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_POST['stage'])</warning>;
    }
}

function deferredClosure()
{
    try {
        return function () {
            return <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_GET['stage'])</warning>;
        };
    } catch (ValueError $e) {
        return null;
    }
}

function deferredArrow()
{
    try {
        return fn () => <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_COOKIE['stage'])</warning>;
    } catch (Throwable $e) {
        return null;
    }
}

function deferredAfterGuard()
{
    if (in_array($_GET['stage'], ['draft', 'live'], true)) {
        return fn () => <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_GET['stage'])</warning>;
    }
    return null;
}

function checkFollowedByCall()
{
    if (in_array($_GET['stage'], ["draft", "live"], true) && audit()) {
        return <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_GET['stage'])</warning>;
    }
    return null;
}

function unknownLiteral()
{
    if (isset($_GET['stage']) && in_array($_GET['stage'], ["draft", "gone"], true)) {
        return <warning descr="Handle unknown enum backing values explicitly.">Stage::from($_GET['stage'])</warning>;
    }
    return null;
}
