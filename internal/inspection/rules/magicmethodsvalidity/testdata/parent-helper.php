<?php
abstract class Store {
    public function __construct(string $name) { echo $name; }
}
// The parent constructor runs from a helper no subclass can replace.
class PrivateHelper extends Store {
    public function __construct(string $dsn) { $this->init($dsn); }
    private function init(string $dsn): void { parent::__construct($dsn); }
}
trait Connects {
    private function connect(string $dsn): void { parent::__construct($dsn); }
}
class TraitHelper extends Store {
    use Connects;
    public function __construct(string $dsn) { $this->connect($dsn); }
}
class SelfHelper extends Store {
    public function __construct(string $dsn) { self::boot($dsn); }
    private function boot(string $dsn): void { parent::__construct($dsn); }
}
final class FinalClass extends Store {
    public function __construct(string $dsn) { $this->init($dsn); }
    protected function init(string $dsn): void { parent::__construct($dsn); }
}
// A subclass may replace the helper, or the helper calls something else.
class Overridable extends Store {
    public function <error descr="__construct does not call Store::__construct().">__construct</error>(string $dsn) { $this->init($dsn); }
    protected function init(string $dsn): void { parent::__construct($dsn); }
}
class OtherCall extends Store {
    public function <error descr="__construct does not call Store::__construct().">__construct</error>(string $dsn) { $this->init($dsn); }
    private function init(string $dsn): void { parent::reset($dsn); }
}
// Only a call on the object itself runs its helper; an anonymous class's
// helper is not looked up.
class OtherObject extends Store {
    public function <error descr="__construct does not call Store::__construct().">__construct</error>(PrivateHelper $other, string $dsn) { $other->init($dsn); }
    private function init(string $dsn): void { parent::__construct($dsn); }
}
class OtherClass extends Store {
    public function <error descr="__construct does not call Store::__construct().">__construct</error>(string $dsn) { SelfHelper::boot($dsn); }
    private static function boot(string $dsn): void {}
}
$anonymous = new class('x') extends Store {
    public function <error descr="__construct does not call Store::__construct().">__construct</error>(string $dsn) { $this->init($dsn); }
    private function init(string $dsn): void { parent::__construct($dsn); }
};
