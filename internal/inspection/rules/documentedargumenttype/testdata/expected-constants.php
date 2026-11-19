<?php

use JetBrains\PhpStorm\ExpectedValues;

class Mode
{
    const FAST = 1;
    const SLOW = 2;
}

class Label
{
    const ON = 'on';
    const OFF = 'off';
}

class Computed
{
    const BASE = 1;
    const NEXT = self::BASE + 1;
}

class Extended extends Mode
{
    const OWN = 5;
}

function listed(#[ExpectedValues([Mode::FAST, 4])] int $mode) {}

function fromClass(#[ExpectedValues(valuesFromClass: Mode::class)] int $mode) {}

function labels(#[ExpectedValues(valuesFromClass: Label::class)] string $label) {}

function flags(#[ExpectedValues(flags: [Mode::FAST, Mode::SLOW])] int $mask) {}

function flagsFromClass(#[ExpectedValues(flagsFromClass: Mode::class)] int $mask) {}

function optional(#[ExpectedValues([Mode::FAST])] ?int $mode = null) {}

function implicit(#[ExpectedValues([Mode::FAST])] int $mode = null) {}

function computed(#[ExpectedValues(valuesFromClass: Computed::class)] int $value) {}

function inherited(#[ExpectedValues(valuesFromClass: Extended::class)] int $value) {}

function missing(#[ExpectedValues(valuesFromClass: Nowhere::class)] int $value) {}

function unknownConstant(#[ExpectedValues([Mode::NONE])] int $value) {}

function calls(int $any)
{
    listed(<warning descr="Pass Mode::FAST|4 to $mode of listed(), not 2.">2</warning>);
    fromClass(<warning descr="Pass Mode::* to $mode of fromClass(), not 9.">9</warning>);
    labels(<warning descr="Pass Label::* to $label of labels(), not 'maybe'.">'maybe'</warning>);
    flags(<warning descr="Pass int-mask<Mode::FAST,Mode::SLOW> to $mask of flags(), not 4.">4</warning>);
    flagsFromClass(<warning descr="Pass int-mask-of<Mode::*> to $mask of flagsFromClass(), not 8.">8</warning>);
    optional(<warning descr="Pass Mode::FAST|null to $mode of optional(), not 2.">2</warning>);

    listed(Mode::FAST);
    listed(1);
    listed(4);
    listed($any);
    fromClass(2);
    labels('off');
    flags(0);
    flags(3);
    flagsFromClass(3);
    optional(null);
    optional();
    implicit(null);
    computed(9);
    inherited(9);
    missing(9);
    unknownConstant(9);
}
