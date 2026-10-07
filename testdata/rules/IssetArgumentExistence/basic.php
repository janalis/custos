<?php
class Billing
{
    public function total($items)
    {
        $sum = <error descr="Variable '$sum' is not defined in this scope.">$sum</error> ?? 0;
        $hit = isset(<error descr="Variable '$cache' is not defined in this scope.">$cache</error>, $items);
        $nil = empty(<error descr="Variable '$nil' is not defined in this scope.">$nil</error>);
        $both = isset(<error descr="Variable '$q' is not defined in this scope.">$q</error>) && empty($q);
        return [$sum, $hit, $nil, $both];
    }

    public function including()
    {
        include 'defaults.php';
        return isset(<error descr="Variable '$config' is not defined in this scope.">$config</error>);
    }
}

function chained()
{
    return <error descr="Variable '$a' is not defined in this scope.">$a</error> ?? <error descr="Variable '$b' is not defined in this scope.">$b</error> ?? 1;
}
