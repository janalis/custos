<?php
class Dispatcher
{
    public function render(array $displayData, string $path): void
    {
        extract($displayData);
        unset($displayData);
        include $path;
    }

    public function vars(array $data): array
    {
        unset($data);
        return get_defined_vars();
    }

    public function packed(array $data, $title): array
    {
        unset($data);
        return compact('title');
    }

    public function evaluated(string $code, $tmp): void
    {
        unset($tmp);
        eval($code);
    }

    public function plain($tmp): void
    {
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($tmp);</weak_warning>
        $f = function () { include 'x.php'; };
    }
}
