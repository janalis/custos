<?php
class Suite
{
    protected function <weak_warning descr="Declare ': void' as the return type.">skip</weak_warning>($msg = '')
    {
        throw new RuntimeException($msg);
    }
}
class BaseCase extends Suite
{
    public function testRun()
    {
        echo 'run';
    }

    public function <weak_warning descr="Declare ': void' as the return type (update the whole hierarchy with a signature refactoring).">testStop</weak_warning>()
    {
        echo 'stop';
    }
}
class SkippedCase extends BaseCase
{
    public function testRun()
    {
        return $this->skip();
    }

    public function <weak_warning descr="Declare ': void' as the return type (update the whole hierarchy with a signature refactoring).">testStop</weak_warning>()
    {
        $f = function () { return 1; };
        return;
    }
}
abstract class Named
{
    /** @return string */
    abstract public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">name</weak_warning>();
}
class Unnamed extends Named
{
    public function name()
    {
        throw new LogicException('none');
    }
}
