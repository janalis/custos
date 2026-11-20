<?php // presets: per psr12 symfony laravel drupal psr1
namespace App\Http;
use Zed\Alpha;
use function strlen;
use App\Beta as B;
#[Attr( 1 )]
abstract class Foo extends Bar implements Baz,Qux {
    use TraitA;
    const A=1;
    const B = 2;
    public  $x=array(1,2,3);
    private ?int $y = NULL;
    protected static $z;
    public function __construct(private readonly int $id, protected string $name = 'x') {}
    abstract public function bar(): void;
    public function foo(INT $a, $b=null, &$c = [], int ...$rest) : ?STRING {
        if($a==1){ return 'a'.$b; }else if ($a){
        foo( $a,$b );
        }elseif($b) {
            $c = (Integer)$a;
        } else {
            $c = - $a;
        }
        $f = function($x) use($a){return $x+$a;};
        $g = fn ($x) => $x*2;
        $h = static fn($x): int => $x;
        foreach($this->x as $k=>&$v){
          echo $k, $v;
        }
        for($i=0;$i<10;$i++){ continue; }
        while ( $a-- ) { break; }
        do { $a++; } while($a < 3);
        switch($a){
            case 1:
                $x = 1;
                break;
            case 2: {
                $x = 2;
            }
            default:
                return (int)$a;
        }
        try {
            $r = $a ?: $b;
        } catch(\Exception | \Error $e) {
            throw $e;
        } finally {
            $r = $a ?? $b;
        }
        $m = match($a) { 1, 2 => 'low', default => 'high' };
        $s = $this?->foo['x'] -> bar(...) :: BAZ;
        list($p, $q) = [1, 2];
        [$p, , $q] = $arr;
        $o = new Foo;
        $n = new class(1) extends Foo {};
        $str = "x {$a} y";
        unset($a, $b);
        return !$a && $b instanceof Foo || $c <> 1;
    }
}
enum Suit: string {
    case Hearts = 'H';
    case Spades='S';
}
interface I {}
trait T { public function t() { return __CLASS__ . __line__; } }
function free(Array $a = Array(), $b = TRUE) { goto end; end: return FALSE; }
$x = $cond
    ? $a
    : $b;
$long = foo($alpha, $beta, $gamma, $delta, $epsilon, $zeta, $eta, $theta, $iota, $kappa, $lambda);
$arr = ['a' => 1,
    'b' => 2];
echo "done";
?>
