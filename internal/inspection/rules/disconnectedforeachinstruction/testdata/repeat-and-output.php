<?php
function attempts(array $matches, $customer, array $rows)
{
    $lines = [];
    foreach ($matches as $key => $val) {
        var_dump($matches);
        print_r($matches);
        $lines[] = $val[0];
    }
    foreach (range(1, 10) as $attempt) {
        post('/login', ['email' => $customer->email]);
        check($customer);
    }
    foreach ($rows as $row) {
        $name = 'row';
        echo compact('row', 'name');
    }
    foreach ($rows as $row) {
        $n = 'row';
        echo $$n;
    }
    foreach ($rows as $row) {
        include 'partial.php';
    }
    foreach ($rows as $row) {
        eval('return $row;');
    }
    foreach ($rows as [$a, $b]) {
        <weak_warning descr="Statement does not depend on the loop; move it out.">check($customer);</weak_warning>
        consume($b);
    }
    return $lines;
}
