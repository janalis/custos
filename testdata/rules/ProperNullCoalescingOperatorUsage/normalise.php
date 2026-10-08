<?php
trait Helper
{
    public function up(): ?parent { return null; }

    public function fallback()
    {
        return $this->up() ?? 'none';
    }
}

class Plain
{
    public function me(): ?self { return null; }

    public function check(?Closure $c, ?bool $flag)
    {
        return [
            $this ?? 'x',
            <weak_warning descr="Operand types of '??' do not match ([callable] vs [string]).">$c ?? 'strlen'</weak_warning>,
            $c ?? fn() => 1,
            $flag ?? true,
            $flag ?? 0,                     // scalar-only sides (D5a)
        ];
    }
}

class NoParent
{
    public function up(): ?parent { return null; }

    public function fallback()
    {
        return $this->up() ?? 'none';
    }
}
