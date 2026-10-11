<?php
function logged_only(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if ($position !== strlen($text)) {
        error_log("Unexpected text after the number");
    }
    <warning descr="Reject unparsed number suffixes.">echo $amount</warning>;
}

function wrong_branch(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $formatter->getErrorCode();
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if ($position === strlen($text)) {
        error_log("Whole input parsed");
    } else {
        <warning descr="Reject unparsed number suffixes.">echo $amount</warning>;
    }
}

function either_condition(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if ($amount !== false || $position === strlen($text)) {
        <warning descr="Reject unparsed number suffixes.">echo $amount</warning>;
    }
}
