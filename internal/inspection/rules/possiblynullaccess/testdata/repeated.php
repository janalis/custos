<?php
namespace Shop;

class Item
{
    public const KIND = 'item';
    public static int $made = 0;
    public int $qty = 0;
    public ?Item $next = null;

    public function price(): int { return 1; }
    public function rate(int $n): int { return $n; }
    public static function make(): string { return ''; }
}

class Repeat
{
    private ?Item $last = null;

    public function find(): ?Item { return null; }
    public function take(?Item $item): void {}

    // Accesses that run only after the first one on the same value.
    public function straight(?Item $x, ?Item $y, ?Item $z, array $list): array
    {
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$x</warning>->price();
        $x->price();
        echo $x->qty;
        echo $x::KIND . $x::$made . $x::make();
        if ($x->price() > 0) {
            echo $x->qty;
        }
        foreach ($list as $i) {
            echo $x->qty;
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->last</warning>->price();
        $this->refresh();
        echo $this->last->qty;
        while (<warning descr="Check this value for null before the access: it is Shop\Item|null.">$z</warning>->qty > 0) {
            $z->price();
            break;
        }
        return [<warning descr="Check this value for null before the access: it is Shop\Item|null.">$y</warning>->qty, $y->price(), (int) $y->qty . 'x', $y?->qty, "{$y->qty}"];
    }

    // The first access does not run on every path to the second, or the
    // value may have changed between them: both are reported.
    public function separate(?Item $a, ?Item $b, ?Item $c, ?Item $d, ?Item $e, ?Item $f, ?Item $g, ?Item $h, bool $ok, array $list): void
    {
        if ($ok) {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$a</warning>->price();
        } else {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$a</warning>->price();
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$a</warning>->price();
        $ok && <warning descr="Check this value for null before the access: it is Shop\Item|null.">$b</warning>->price();
        $ok ? <warning descr="Check this value for null before the access: it is Shop\Item|null.">$b</warning>->price() : 0;
        echo <warning descr="Check this value for null before the access: it is Shop\Item|null.">$b</warning>->qty;
        foreach ($list as $i) {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$c</warning>->price();
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$c</warning>->price();
        try {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$d</warning>->price();
        } catch (\Error $error) {
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$d</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$e</warning>->price();
        $e = $this->find();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$e</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$f</warning>->price();
        $this->take($f);
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$f</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$g</warning>->price();
        foreach ($list as $i) {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$g</warning>->price();
            $g = $this->find();
        }
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$h</warning>->rate(<warning descr="Check this value for null before the access: it is Shop\Item|null.">$h</warning>->qty);
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->last</warning>->price();
        $this->last = $this->find();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->last</warning>->price();
    }

    // Known limit: as a check of a property outlives calls, so does an
    // access on it. reset() setting it to null in between is not seen.
    public function survives(): void
    {
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$this->last</warning>->price();
        $this->reset();
        $this->last->price();
    }

    public function reset(): void { $this->last = null; }

    // A `foreach` subject or a `for` initializer runs once, before the
    // first iteration: a write later in the loop reaches the next one.
    public function once(?Item $a, ?Item $b, ?Item $c): void
    {
        foreach ([<warning descr="Check this value for null before the access: it is Shop\Item|null.">$a</warning>->qty] as $i) {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$a</warning>->price();
            $a = $this->find();
        }
        for ($n = <warning descr="Check this value for null before the access: it is Shop\Item|null.">$b</warning>->qty; $n > 0; $n--) {
            <warning descr="Check this value for null before the access: it is Shop\Item|null.">$b</warning>->price();
            $b = $this->find();
        }
        foreach ([<warning descr="Check this value for null before the access: it is Shop\Item|null.">$c</warning>->qty] as $i) {
            $c->price();
        }
    }

    // Elements: a read repeats a read; a write turns null into an array.
    public function elements(?array $rows, ?array $more): void
    {
        echo <warning descr="Check this value for null before the access: it is array|null.">$rows</warning>['a'];
        echo $rows['b'];
        echo isset($more['a']) ? 1 : 0;
        echo <warning descr="Check this value for null before the access: it is array|null.">$more</warning>['b'];
        $more['c'] = 1;
        echo $more['d'];
    }

    // A property of another object is no value told apart.
    public function chained(Item $item): void
    {
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item->next</warning>->price();
        <warning descr="Check this value for null before the access: it is Shop\Item|null.">$item->next</warning>->price();
    }

    public function refresh(): void {}
}

$top = (new Repeat())->find();
<warning descr="Check this value for null before the access: it is Shop\Item|null.">$top</warning>->price();
$top->price();

// A `goto` may jump past the earlier access.
function jumps(?Item $x, bool $skip): void
{
    if ($skip) {
        goto later;
    }
    <warning descr="Check this value for null before the access: it is Shop\Item|null.">$x</warning>->price();
    later:
    <warning descr="Check this value for null before the access: it is Shop\Item|null.">$x</warning>->price();
}

// An earlier access whose finding is suppressed reports nothing: the next
// access is reported in its place.
function silenced(?Item $x, ?Item $y, ?Item $z): void
{
    /** @custos-ignore PossiblyNullAccess */
    $x->price();
    <warning descr="Check this value for null before the access: it is Shop\Item|null.">$x</warning>->price();
    $x->price();
    <warning descr="Check this value for null before the access: it is Shop\Item|null.">$y</warning>->price();
    // @noinspection PossiblyNullAccess
    $y->price();
    $y->price();
    // @noinspection NullPointerException
    $z->price();
    <warning descr="Check this value for null before the access: it is Shop\Item|null.">$z</warning>->price();
}
