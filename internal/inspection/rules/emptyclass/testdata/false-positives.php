<?php
class HasState  { private $count = 0; }
class HasLimit  { const MAX = 10; }
class HasAction { public function go() {} }
trait Greets    { public function hi() {} }
class UsesTrait { use Greets; }
class Promoted  { public function __construct(private int $x) {} }
interface Marker {}
enum Size { case Small; }

class ParseFailure extends \InvalidArgumentException {}
class DeepFailure extends ParseFailure {}
abstract class Shape { abstract public function area(); }
class Dot extends Shape {}
/** @deprecated use HasState */
class OldStub {}
$x = new class {};
