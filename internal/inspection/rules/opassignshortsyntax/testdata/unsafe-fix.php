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
        <weak_warning descr="Use the compound form '$this-&gt;out .= $this-&gt;child()'.">$this->out = $this->out . $this->child()</weak_warning>;
        <weak_warning descr="Use the compound form '$this-&gt;parts['n'] += bump($this)'.">$this->parts['n'] = $this->parts['n'] + bump($this)</weak_warning>;
        <weak_warning descr="Use the compound form 'self::$hits += self::touch()'.">self::$hits = self::$hits + self::touch()</weak_warning>;
        <weak_warning descr="Use the compound form '$this-&gt;parts[$k] -= $k++'.">$this->parts[$k] = $this->parts[$k] - $k++</weak_warning>;
        <weak_warning descr="Use the compound form '$$name .= new Part()'.">$$name = $$name . new Part()</weak_warning>;
        <weak_warning descr="Use the compound form '$this-&gt;out .= 'a' . ($this-&gt;out = 'b')'.">$this->out = $this->out . 'a' . ($this->out = 'b')</weak_warning>;

        // Without a possible write in the fragment the fix stays.
        <weak_warning descr="Use the compound form '$this-&gt;out .= $name . '!''.">$this->out = $this->out . $name . '!'</weak_warning>;
        <weak_warning descr="Use the compound form 'self::$hits += $k'.">self::$hits = self::$hits + $k</weak_warning>;
        <weak_warning descr="Use the compound form '$this-&gt;parts['f'] .= fn () =&gt; bump($this)'.">$this->parts['f'] = $this->parts['f'] . fn () => bump($this)</weak_warning>;
    }
}

function totals($x, $n, $s, $k)
{
    // '+' and '*' chains regroup differently for floats and on overflow.
    <weak_warning descr="Use the compound form '$x += 1 + 1'.">$x = $x + 1 + 1</weak_warning>;
    <weak_warning descr="Use the compound form '$n *= PHP_INT_MAX * 2'.">$n = $n * PHP_INT_MAX * 2</weak_warning>;

    // A plain variable is read late in both forms, except by the fragments
    // that follow the first one of a chain.
    <weak_warning descr="Use the compound form '$s .= 'a' . tail()'.">$s = $s . 'a' . tail()</weak_warning>;
    <weak_warning descr="Use the compound form '$s .= 'a' . ($s = 'b')'.">$s = $s . 'a' . ($s = 'b')</weak_warning>;
    <weak_warning descr="Use the compound form '$s .= tail() . 'a''.">$s = $s . tail() . 'a'</weak_warning>;
    <weak_warning descr="Use the compound form '$n += tail()'.">$n = $n + tail()</weak_warning>;
    <weak_warning descr="Use the compound form '$n -= $k++'.">$n = $n - $k++</weak_warning>;
}
