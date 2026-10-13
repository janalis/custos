<?php
// Reported, but the compound form is not provably the same value: no fix.
class View
{
    public $out = '';
    public $parts = ['n' => 1];
    public static $hits = 0;

    function render($name, $k)
    {
        // The long form reads the property or element before the call runs.
        $this->out = $this->out . $this->child();
        $this->parts['n'] = $this->parts['n'] + bump($this);
        self::$hits = self::$hits + self::touch();
        $this->parts[$k] = $this->parts[$k] - $k++;
        $$name = $$name . new Part();
        $this->out = $this->out . 'a' . ($this->out = 'b');

        // Without a possible write in the fragment the fix stays.
        $this->out .= $name . '!';
        self::$hits += $k;
        $this->parts['f'] .= fn () => bump($this);
    }
}

function totals($x, $n, $s, $k)
{
    // '+' and '*' chains regroup differently for floats and on overflow.
    $x = $x + 1 + 1;
    $n = $n * PHP_INT_MAX * 2;

    // A plain variable is read late in both forms, except by the fragments
    // that follow the first one of a chain.
    $s = $s . 'a' . tail();
    $s = $s . 'a' . ($s = 'b');
    $s .= tail() . 'a';
    $n += tail();
    $n -= $k++;
}
