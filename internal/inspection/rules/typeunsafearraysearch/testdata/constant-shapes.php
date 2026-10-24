<?php
class Telegram
{
    const MODES = ['HTML', 'Markdown'];
    const MIXED = ['HTML', 1];
    const NONE = [];
    const CODES = [1, 2];

    public function check(?string $mode, string $s, int $n, $any)
    {
        // Every value of the constant is a string, like the needle.
        if ($mode !== null && !in_array($mode, self::MODES)) {
            return false;
        }
        in_array($n, self::CODES);
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($s, self::MIXED)</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($s, self::NONE)</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($n, self::MODES)</weak_warning>;
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($any, self::MODES)</weak_warning>;
        return true;
    }
}

trait Aggregate
{
    /** @var string[] */
    private static $values = ['SUM', 'MIN'];
    private static $plain = ['A', 'B'];

    public function accepts($argument)
    {
        // The trait's own property keeps its type in the using class.
        if (is_string($argument) && in_array(strtoupper($argument), self::$values)) {
            return true;
        }
        return is_string($argument) && <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($argument, self::$plain)</weak_warning>;
    }
}
