<?php
class Menu
{
    private $menu = [];

    public function edit($main, $sub)
    {
        // A reference into the array is kept: no '??' rewrite.
        if (isset($this->menu[$main])) {
            $item = &$this->menu[$main];
        } else {
            $item = null;
        }
        $label = null;
        if (isset($this->menu[$sub])) {
            $label = &$this->menu[$sub];
        }
        $item['edited'] = true;
        $label = 'x';
    }

    public function &entry($key)
    {
        if (isset($this->menu[$key])) {
            return $this->menu[$key];
        }
        return $this->menu;
    }

    public function plain($key)
    {
        <weak_warning descr="Simplify to 'return $this->menu[$key] ?? null' using the null coalescing operator.">if</weak_warning> (isset($this->menu[$key])) {
            return $this->menu[$key];
        }
        return null;
    }
}

$sum = function &($k) use (&$data) {
    if (isset($data[$k])) {
        return $data[$k];
    }
    return $data;
};

function &lookup(array &$data, $k)
{
    if (isset($data[$k])) {
        return $data[$k];
    }
    return $data;
}
