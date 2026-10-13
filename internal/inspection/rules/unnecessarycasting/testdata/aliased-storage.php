<?php
function fillQuantity(&$out) { $out = '5 pears'; }
function showQuantity($value, string $unit = '') { echo $value, $unit; }
function fillAll($first, &...$rest) { $rest[0] = '4 plums'; }

class Stock
{
    public function __construct($start = 0, &$report = null) { $report = 'ready'; }
    public function fill(&$slot) { $slot = '6 figs'; }
    public static function fillShared(&$slot) { $slot = '7 limes'; }
    public function peek($slot) { return $slot; }
    public static function peekShared($slot) { return $slot; }

    public function viaMethod() { $n = 0; $this->fill($n); return (int) $n; }
    public function viaStatic() { $n = 0; self::fillShared($n); return (int) $n; }
    public function viaCallable() { $n = 0; $f = [$this, 'fill']; $f($n); return (int) $n; }
    public function byValueMethod() { $n = 0; $this->peek($n); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
    public function byValueStatic() { $n = 0; self::peekShared($n); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
}

function viaFunction() { $n = 0; fillQuantity($n); return (int) $n; }
function viaUnknownFunction() { $n = 0; not_declared_anywhere($n); return (int) $n; }
function viaUnknownReceiver($service) { $n = 0; $service->load($n); return (int) $n; }
function viaUnknownClass() { $n = 0; $o = new NotDeclaredAnywhere($n); return (int) $n; }
function viaConstructor() { $n = 0; $o = new Stock(1, $n); return (int) $n; }
function viaVariadic() { $n = 0; fillAll(1, 2, $n); return (int) $n; }
function viaNamed() { $n = 0; fillQuantity(out: $n); return (int) $n; }
function viaNamedVariadic() { $n = 0; fillAll(1, extra: $n); return (int) $n; }
function viaSpread() { $n = [1]; showQuantity(...$n); return (array) $n; }
function viaScanf() { $n = 0; sscanf('abc', '%s', $n); return (int) $n; }
function viaSettype() { $n = 0; settype($n, 'string'); return (int) $n; }
function viaElement($in) { $a = [1, 2]; array_unshift($a, $in); return (int) $a[0]; }
function viaNestedElement() { $a = ['k' => ['j' => 1]]; fillQuantity($a['k']['j']); return (int) $a['k']['j']; }

function byValueFunction() { $n = 0; showQuantity($n); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function byValueNamed() { $n = 0; showQuantity(unit: 'kg', value: $n); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function byValueConstructor() { $n = 0; $o = new Stock($n); $p = new Stock; return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function byValueFirst() { $n = 0; $m = 0; fillAll($n, $m); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function byValueExtra() { $n = 0; showQuantity(1, 'kg', $n); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function firstClassCallable() { $n = 0; $f = showQuantity(...); return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }

function aliasOf(int $n) { $r = &$n; $r = '7 apples'; return (int) $n; }
function aliasTo(array $box) { $n = 0; $n = &$box['k']; $box['k'] = 'x'; return (int) $n; }
function aliasOfElement($in) { $a = ['k' => 1]; $b = &$a['k']; $b = $in; return (int) $a['k']; }
function aliasInArray() { $n = 0; $refs = [&$n]; $refs[0] = 'x'; return (int) $n; }
function loopByReference($in) { $row = ['qty' => 1]; foreach ($row as &$cell) { $cell = $in; } return (int) $row['qty']; }
function loopBinds(array $rows) { $n = 0; foreach ($rows as &$n) { } return (int) $n; }
function loopDestructures($in) { $rows = [[1]]; foreach ($rows as [&$cell]) { $cell = $in; } return (array) $rows; }
function loopByValue() { $rows = [[1]]; foreach ($rows as [$cell]) { echo $cell; } return <weak_warning descr="Operand already has the target type; remove the cast.">(array)</weak_warning> $rows; }
function loopOverOther(array $other) { $n = 0; foreach ($other as &$cell) { $cell = 1; } foreach ($other as $cell) { } return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function aliasThenLoop($in) { $n = 0; $r = &$n; $r = $in; foreach ([1] as $x) { } return (int) $n; }
function passedThenLoop() { $n = 0; fillQuantity($n); foreach ([1] as $x) { } return (int) $n; }
function passedThenSibling() { $n = 0; $z = [fillQuantity($n), strlen('x')]; return (int) $n; }
function aliasThenSibling(int $n) { $z = [$r = &$n, $q = 1]; $r = 'x'; return (int) $n; }
function declaredGlobal() { global $shared, $n; $n = 0; refreshShared(); return (int) $n; }
function insideArrow(int $n) { fillQuantity($n); return fn() => fn() => (int) $n; }
function arrowParameter() { return fn(int $n) => <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n; }
function viaUnknownStatic() { $n = 0; NotDeclaredAnywhere::load($n); return (int) $n; }
function arrowPasses() { $n = 0; $f = fn() => fillQuantity($n); return (int) $n; }

function otherScopes()
{
    $n = 0;
    $inner = function () { $n = 1; fillQuantity($n); };
    $object = new class { public function run() { $n = 1; fillQuantity($n); } };
    return <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n;
}
