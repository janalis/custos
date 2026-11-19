<?php

/** @template T */
class Box {}

/** @template-covariant T */
class Source {}

/** @template-contravariant T */
class Sink {}

/** @param Box<int> $box */
function takesBox(Box $box) {}

/** @param Source<int> $source */
function takesSource(Source $source) {}

/** @param Sink<int> $sink */
function takesSink(Sink $sink) {}

/** @param Box<int, string> $box */
function takesTwoArguments(Box $box) {}

/** @param Missing<int> $missing */
function takesMissing($missing) {}

/**
 * @param Box<string> $box
 * @param Source<string> $source
 * @param Sink<string> $sink
 */
function mismatches(Box $box, Source $source, Sink $sink)
{
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$box</warning>);
    takesSource(<warning descr="Pass Source<int> to $source of takesSource(), not Source<string>.">$source</warning>);
    takesSink(<warning descr="Pass Sink<int> to $sink of takesSink(), not Sink<string>.">$sink</warning>);
}

/**
 * @param Box<int> $box
 * @param Box<int|string> $wide
 * @param Box $raw
 * @param Box<int, int> $extra
 * @param Missing<string> $missing
 */
function fits(Box $box, Box $wide, Box $raw, Box $extra, $missing)
{
    takesBox($box);
    takesBox($wide);
    takesBox($raw);
    takesTwoArguments($extra);
    takesMissing($missing);
}

/** @return Box<string> */
function makeBox() {}

/**
 * @template T
 * @param T $value
 * @return Box<T>
 */
function boxOf($value) {}

/** @extends Box<string> */
class StringBox extends Box {}

class DeepStringBox extends StringBox {}

/**
 * @template U
 * @extends Box<U>
 */
class Wrapper extends Box {}

/** @implements Iterator<int, string> */
abstract class Lines implements Iterator {}

class Factory
{
    /** @var Box<string> */
    public $stored;

    /** @return Box<string> */
    public function make() {}

    /** @return Box<string> */
    public static function build() {}

    public function run()
    {
        takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$this->stored</warning>);
        takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$this->make()</warning>);
        takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">self::build()</warning>);
    }
}

/**
 * @template T
 */
class Holder
{
    /** @var Box<T> */
    public $held;

    /** @return Box<T> */
    public function get() {}

    public function run()
    {
        takesBox($this->held);
        takesBox($this->get());
    }
}

function declaredSources(Factory $factory)
{
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">makeBox()</warning>);
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$factory->make()</warning>);
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">Factory::build()</warning>);
    $local = makeBox();
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$local</warning>);
    $again = $local;
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Box<string>.">$again</warning>);
}

/**
 * @param StringBox $sub
 * @param DeepStringBox $deep
 * @param Wrapper<string> $wrapped
 * @param Wrapper $raw
 */
function subclasses(StringBox $sub, DeepStringBox $deep, Wrapper $wrapped, Wrapper $raw)
{
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not StringBox.">$sub</warning>);
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not DeepStringBox.">$deep</warning>);
    takesBox(<warning descr="Pass Box<int> to $box of takesBox(), not Wrapper<string>.">$wrapped</warning>);
    takesBox($raw);
}

/** @param Box<int> $box */
function notDeclared(bool $flag, Factory $factory, Box $box, $unknown)
{
    takesBox(boxOf('text'));
    $picked = $flag ? makeBox() : $box;
    takesBox($picked);
    if ($flag) {
        $branch = makeBox();
    } else {
        $branch = $box;
    }
    takesBox($branch);
    takesBox($unknown->make());
    takesBox($factory->missing());
    takesBox($factory->{$unknown}());
    takesBox($unknown::build());
    takesBox(Missing::build());
    takesBox($factory->stored);
    takesBox(missingFunction());
}
