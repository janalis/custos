<?php
// The pattern consumes the unit, so nothing is left unparsed.
$weight = new NumberFormatter("en_US", NumberFormatter::PATTERN_DECIMAL, "#0'kg'");
$kilos = $weight->parse("37kg", NumberFormatter::TYPE_DOUBLE, $kiloEnd);
echo $kilos;

// A configured suffix is part of the accepted number.
$suffixed = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
$suffixed->setTextAttribute(NumberFormatter::POSITIVE_SUFFIX, "kg");
$grams = $suffixed->parse("12kg", NumberFormatter::TYPE_DOUBLE, $gramEnd);
echo $grams;

// Unknown styles and reconfigured formatters prove nothing.
function custom_style(int $style): void
{
    $formatter = new NumberFormatter("en_US", $style);
    $amount = $formatter->parse("5usd");
    echo $amount;
}

function early_return(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if ($amount === false || $position !== strlen($text)) {
        return;
    }
    echo $amount;
}

function early_throw(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if (strlen($text) !== $position) {
        throw new InvalidArgumentException("Unexpected text after the number");
    }
    if ($amount !== false) {
        echo $amount;
    }
}

function else_branch(): void
{
    $formatter = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
    $text = "37tail";
    $amount = $formatter->parse($text, NumberFormatter::TYPE_DOUBLE, $position);
    if ($position !== strlen($text)) {
        error_log("Unexpected text after the number");
    } else {
        echo $amount;
    }
}

$strict = new NumberFormatter("en_US", NumberFormatter::DECIMAL);
$raw = "37tail";
$parsed = $strict->parse($raw, NumberFormatter::TYPE_DOUBLE, $stop);
if ($stop !== strlen($raw)) {
    exit("Unexpected text after the number\n");
}
echo $parsed;
