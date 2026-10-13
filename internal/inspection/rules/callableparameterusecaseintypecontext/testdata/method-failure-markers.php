<?php
class Alias
{
    /** @return string|false */
    public static function get(string $a) { return $a === '' ? false : $a; }

    public function resize(string $k): ?string { return $k === '' ? null : $k; }

    public function useThem(string $path, string $key, ?Alias $o): void
    {
        $path = Alias::get($path);
        $key = $this->resize($key);
        $key = $o?->resize($key);
        $path = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$path === 'x'</warning>;
    }
}
