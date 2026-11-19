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

const LEVEL_LOW = 1;
const LEVEL_HIGH = 2;
const LEVEL_COMPUTED = LEVEL_LOW + 1;

class Base
{
    const PARENT_VALUE = 7;
}

class Owner extends Base
{
    const OWN_VALUE = 3;

    public function own(#[ExpectedValues([self::OWN_VALUE])] int $value) {}

    public function late(#[ExpectedValues([static::OWN_VALUE])] int $value) {}

    public function inherited(#[ExpectedValues([parent::PARENT_VALUE])] int $value) {}
}

class Orphan
{
    public function noParent(#[ExpectedValues([parent::NOTHING])] int $value) {}
}

function levels(#[ExpectedValues([LEVEL_LOW, LEVEL_HIGH])] int $level) {}

function computedLevel(#[ExpectedValues([LEVEL_COMPUTED])] int $level) {}

function noSuchLevel(#[ExpectedValues([LEVEL_MISSING])] int $level) {}

function keyword(#[ExpectedValues([null])] ?int $level) {}

class Letters
{
    const A = 'a';
}

/** @param int $value */
function mixedKinds(#[ExpectedValues([Letters::A, 'x'])] $value) {}

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

    levels(<warning descr="Pass LEVEL_LOW|LEVEL_HIGH to $level of levels(), not 3.">3</warning>);
    (new Owner())->own(<warning descr="Pass Owner::OWN_VALUE to $value of Owner::own(), not 9.">9</warning>);
    (new Owner())->inherited(<warning descr="Pass Base::PARENT_VALUE to $value of Owner::inherited(), not 9.">9</warning>);

    levels(1);
    levels(LEVEL_HIGH);
    (new Owner())->own(3);
    (new Owner())->late(9);
    mixedKinds(9);
    (new Owner())->inherited(7);
    (new Orphan())->noParent(9);
    computedLevel(9);
    noSuchLevel(9);
    keyword(9);
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
    inherited(<warning descr="Pass Extended::* to $value of inherited(), not 9.">9</warning>);
    inherited(5);
    inherited(2);
    missing(9);
    unknownConstant(9);
}
