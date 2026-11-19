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

/** @param Box<string> $box */
function notTheDeclaredArguments(Box $box)
{
    takesBox(makeBox());
    $local = makeBox();
    takesBox($local);
    $box = makeBox();
    takesBox($box);
}
