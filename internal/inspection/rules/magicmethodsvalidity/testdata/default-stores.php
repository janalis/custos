<?php
// A parent constructor that stores defaults into properties already holding them does nothing.
class RepeatsBase
{
    protected array $args = [];
    protected $label;
    public $on = false;
    public $n = 1;
    public function __construct(?string $class = null, array $args = [], $label = null, $on = FALSE, $n = 1)
    {
        if (null !== $class) {
            $this->setClass($class);
        }
        $this->args = $args;
        $this->label = ($label);
        $this->on = $on;
        $this->n = $n;
    }
    public function setClass($c) {}
}
class RepeatsChild extends RepeatsBase
{
    public function __construct() {}
}
class NewValueBase
{
    protected array $args = ['x'];
    public function __construct(array $args = ['x']) { $this->args = $args; }
}
class NewValueChild extends NewValueBase
{
    public function <error descr="__construct does not call NewValueBase::__construct().">__construct</error>() {}
}
class OtherDefaultBase
{
    protected array $args = [];
    public function __construct(array $args = [1]) { $this->args = $args; }
}
class OtherDefaultChild extends OtherDefaultBase
{
    public function <error descr="__construct does not call OtherDefaultBase::__construct().">__construct</error>() {}
}
class TypedUnsetBase
{
    protected ?array $args;
    public function __construct(?array $args = null) { $this->args = $args; }
}
class TypedUnsetChild extends TypedUnsetBase
{
    public function <error descr="__construct does not call TypedUnsetBase::__construct().">__construct</error>() {}
}
class ConstantBase
{
    protected $mode = PHP_EOL;
    public function __construct($mode = PHP_EOL) { $this->mode = $mode; }
}
class ConstantChild extends ConstantBase
{
    public function <error descr="__construct does not call ConstantBase::__construct().">__construct</error>() {}
}
class MixedBase
{
    protected $a = [];
    public function __construct($a = null) { $this->a = $a; }
}
class MixedChild extends MixedBase
{
    public function <error descr="__construct does not call MixedBase::__construct().">__construct</error>() {}
}
class Mixed2Base
{
    protected $a = 1;
    public function __construct($a = []) { $this->a = $a; }
}
class Mixed2Child extends Mixed2Base
{
    public function <error descr="__construct does not call Mixed2Base::__construct().">__construct</error>() {}
}
class Mixed3Base
{
    protected $a = null;
    public function __construct($a = 1) { $this->a = $a; }
}
class Mixed3Child extends Mixed3Base
{
    public function <error descr="__construct does not call Mixed3Base::__construct().">__construct</error>() {}
}
class MissingBase
{
    public function __construct($a = null) { $this->a = $a; }
}
class MissingChild extends MissingBase
{
    public function <error descr="__construct does not call MissingBase::__construct().">__construct</error>() {}
}
class StaticBase
{
    protected static $a = null;
    protected $b;
    public function __construct($a = null) { $this->a = $a; }
}
class StaticChild extends StaticBase
{
    public function <error descr="__construct does not call StaticBase::__construct().">__construct</error>() {}
}
class AppendsBase
{
    protected $a = '';
    public function __construct($a = '') { $this->a .= $a; }
}
class AppendsChild extends AppendsBase
{
    public function <error descr="__construct does not call AppendsBase::__construct().">__construct</error>() {}
}
class OtherBase
{
    protected $a = null;
    public function __construct($a = null) { $that = $this; $that->a = $a; }
}
class OtherChild extends OtherBase
{
    public function <error descr="__construct does not call OtherBase::__construct().">__construct</error>() {}
}
class SafeBase
{
    protected $a = null;
    public function __construct($a = null) { $this?->a = $a; }
}
class SafeChild extends SafeBase
{
    public function <error descr="__construct does not call SafeBase::__construct().">__construct</error>() {}
}
class LocalBase
{
    protected $a = null;
    public function __construct($a = null) { $a = $a; }
}
class LocalChild extends LocalBase
{
    public function <error descr="__construct does not call LocalBase::__construct().">__construct</error>() {}
}
class DynamicBase
{
    protected $a = null;
    public function __construct($a = null) { $this->{'a'} = $a; }
}
class DynamicChild extends DynamicBase
{
    public function <error descr="__construct does not call DynamicBase::__construct().">__construct</error>() {}
}
class LiteralBase
{
    protected $a = null;
    public function __construct($a = null) { $this->a = null; }
}
class LiteralChild extends LiteralBase
{
    public function <error descr="__construct does not call LiteralBase::__construct().">__construct</error>() {}
}
class CallBase
{
    protected $a = null;
    public function __construct($a = null) { strlen($a); }
}
class CallChild extends CallBase
{
    public function <error descr="__construct does not call CallBase::__construct().">__construct</error>() {}
}
