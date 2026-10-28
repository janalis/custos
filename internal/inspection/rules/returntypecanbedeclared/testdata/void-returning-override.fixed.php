<?php
class Suite
{
    protected function skip($msg = ''): void
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

    public function testStop()
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

    public function testStop()
    {
        $f = function () { return 1; };
        return;
    }
}
abstract class Named
{
    /** @return string */
    abstract public function name();
}
class Unnamed extends Named
{
    public function name()
    {
        throw new LogicException('none');
    }
}
