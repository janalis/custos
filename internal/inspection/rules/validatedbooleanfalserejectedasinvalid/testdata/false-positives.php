<?php
// Padding the filter does not strip leaves the input invalid: false means failure here.
function nonBreakingSpace()
{
    if (filter_var("\u{00A0}off", FILTER_VALIDATE_BOOLEAN) === false) {
        return false;
    }
    return true;
}
function emSpace()
{
    if (filter_var("no\u{2003}", FILTER_VALIDATE_BOOLEAN) === false) {
        return false;
    }
    return true;
}
function formFeed()
{
    if (filter_var("\ffalse", FILTER_VALIDATE_BOOLEAN) === false) {
        return false;
    }
    return true;
}
// Array flags reject a scalar or wrap the result, so false is not a validated boolean.
function requireArray()
{
    if (filter_var('false', FILTER_VALIDATE_BOOLEAN, FILTER_REQUIRE_ARRAY) === false) {
        return false;
    }
    return true;
}
function forceArray()
{
    if (filter_var('0', FILTER_VALIDATE_BOOLEAN, FILTER_FORCE_ARRAY) === false) {
        return false;
    }
    return true;
}
