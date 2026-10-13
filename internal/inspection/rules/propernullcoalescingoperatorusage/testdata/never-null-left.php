<?php
namespace App;

class Engine {}

class Gateway {
    public function send(array $data, int $n, ?int $id, mixed $o): array {
        return [
            <weak_warning descr="Operand types of '??' do not match ([bool] vs [int]).">0 < $data['invalid'] ?? 0</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([bool] vs [int]).">1 !== $n ?? 0</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([bool] vs [string]).">$n > 1 && $n < 9 ?? 'no'</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([bool] vs [int]).">$o instanceof Engine ?? 0</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([bool] vs [string]).">!$n ?? 'no'</weak_warning>,
            $id ?? 'new',
            -$n ?? 'none',
            $n + 1 ?? 'none',
        ];
    }
}
