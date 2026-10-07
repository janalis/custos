<?php
class Settings
{
    private $theme;

    public function locale()
    {
        return $this->theme;
    }

    public function reload()
    {
        return parent::reload();
    }

    public function save()
    {
        return $This->save();
    }

    public function sync()
    {
        return ($this->sync());
    }

    public function close($n)
    {
        if ($n > 0) {
            return $this->close($n - 1);
        }
    }

    public function twice()
    {
        $this->log();
        return $this->twice();
    }

    public function sum()
    {
        return 1 + $this->sum();
    }

    public function assign()
    {
        $x = $this->assign();
    }

    public function other()
    {
        return $that->other();
    }

    public function named()
    {
        return Settings::named();
    }

    public function out()
    {
        echo $this->out();
    }
}

interface Shape
{
    public function area();
}

function loop()
{
    return loop();
}

class Proxy
{
    public function dyn($m)
    {
        return $this->$m();
    }

    public function delegate()
    {
        return $this->dyn('x');
    }
}
