<?php
class EmptyBase
{
    public function __construct()
    {
        // nothing to initialise
    }
    public function __clone() {}
}
class FromEmpty extends EmptyBase
{
    public function __construct() { $this->x = 1; }
    public function __clone() { $this->x = 2; }
}
class PromotingBase
{
    public function __construct(protected int $id) {}
}
class FromPromoting extends PromotingBase
{
    public function <error descr="__construct does not call PromotingBase::__construct().">__construct</error>() { $this->id = 1; }
}
class WorkingBase
{
    public function __construct() { $this->ready = true; }
}
class FromWorking extends WorkingBase
{
    public function <error descr="__construct does not call WorkingBase::__construct().">__construct</error>() {}
}
class FromBuiltin extends \ArrayObject
{
    public function <error descr="__construct does not call ArrayObject::__construct().">__construct</error>() {}
}
class CallsWorking extends WorkingBase
{
    private function init(): void {}
    public function __construct()
    {
        $this->init();
        parent::__construct();
        $this->extra = [1, 2];
    }
}
