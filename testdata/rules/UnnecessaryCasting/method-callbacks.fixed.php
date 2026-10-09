<?php
class CallbackLabels {
    public static function length(string $value): int { return strlen($value); }
    public function label(string $value): string { return '[' . $value . ']'; }
    private static function hidden(string $value): int { return strlen($value); }
}

function methodCallbackLabels(CallbackLabels $labels, string $method) {
    echo 'static:' . call_user_func([CallbackLabels::class, 'length'], 'one');
    echo 'string:' . call_user_func('CallbackLabels::length', 'two');
    echo 'object:' . call_user_func([$labels, 'label'], 'three');
    foreach (array_map([$labels, 'label'], ['a', 'b']) as $label) {
        echo 'mapped:' . $label;
    }
    echo 'dynamic:' . (string) call_user_func([$labels, $method], 'four');
    echo 'private:' . (string) call_user_func([CallbackLabels::class, 'hidden'], 'five');
    echo 'instance:' . (string) call_user_func([CallbackLabels::class, 'label'], 'six');
}
