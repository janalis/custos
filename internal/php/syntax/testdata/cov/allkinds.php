<?php
declare(strict_types=1);

namespace App\Demo;

use Foo\Bar as Baz;
use Foo\{Qux, Quux};
use function strlen;
use const PHP_EOL;

const LIMIT = 10, OTHER = 2;

#[Attr(1)]
function gen(int|string $a, (A&B)|null $b = null, ?int ...$rest): iterable
{
    static $calls = 0, $more;
    global $config;
    yield 1;
    yield 'k' => 2;
    yield from [3];
    $x = $a <=> 1;
    $y = -$x + ~1 . "s";
    $x++;
    --$y;
    $z = $x ? $y : null;
    $w = $x ?: $y;
    if ($x instanceof \Countable) {
        echo isset($x[0]), empty($y), PHP_EOL;
    } elseif ($y) {
        print 'p';
    } else {
        unset($z);
    }
    while (false) {
        break;
    }
    do {
        continue;
    } while (false);
    for ($i = 0; $i < LIMIT; $i++) {
    }
    foreach ([1, 'a' => 2, ...$rest] as $k => &$v) {
    }
    switch ($x) {
        case 1:
            break;
        default:
    }
    try {
        throw new \Exception('e');
    } catch (\Exception | \Error $e) {
    } finally {
    }
    [$p, [$q]] = [1, [2]];
    list('a' => $r) = ['a' => 1];
    $f = static function &() use (&$x, $y): void {};
    $g = fn($n) => $n * 2;
    $h = strlen(...);
    $m = match ($x) {
        1, 2 => 'a',
        default => 'b',
    };
    $s = "a $x {$y} ${z}";
    $obj = new \stdClass();
    $obj->prop = $obj?->other;
    $c = Baz::CONST;
    $d = Baz::$static;
    $e2 = Baz::make(name: 1);
    $obj->call(...[1]);
    $anon = new class (1) extends Baz implements \Countable {
        public function count(): int { return 0; }
    };
    $line = __LINE__;
    $cl = clone $obj;
    $inc = include 'x.php';
    $ev = eval('return 1;');
    @$x;
    $paren = (1 + 2);
    goto end;
    end:
    ;
    exit(0);
}

interface I
{
    const C = 1;
    public function m(): static;
}

trait T
{
    public function t(): void {}
}

enum Suit: string implements I
{
    case Hearts = 'h';
    const C = 1;
    public function m(): static { return $this; }
}

abstract class Base
{
    use T {
        t as protected aliased;
        T::t insteadof T;
    }

    public const X = 1;
    protected static ?int $count = null;
    public private(set) int $hooked {
        get => $this->hooked;
        set(int $value) { $this->hooked = $value; }
    }

    public function __construct(private readonly int $id) {}

    abstract protected function run(): never;
}
?>
<b>html</b>
<?php
__halt_compiler();
data
