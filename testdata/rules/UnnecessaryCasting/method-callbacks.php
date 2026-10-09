<?php
class CallbackLabels {
    public static function length(string $value): int { return strlen($value); }
    public function label(string $value): string { return '[' . $value . ']'; }
    private static function hidden(string $value): int { return strlen($value); }
}

function methodCallbackLabels(CallbackLabels $labels, string $method) {
    echo 'static:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> call_user_func([CallbackLabels::class, 'length'], 'one');
    echo 'string:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> call_user_func('CallbackLabels::length', 'two');
    echo 'object:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> call_user_func([$labels, 'label'], 'three');
    foreach (array_map([$labels, 'label'], ['a', 'b']) as $label) {
        echo 'mapped:' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $label;
    }
    echo 'dynamic:' . (string) call_user_func([$labels, $method], 'four');
    echo 'private:' . (string) call_user_func([CallbackLabels::class, 'hidden'], 'five');
    echo 'instance:' . (string) call_user_func([CallbackLabels::class, 'label'], 'six');
}
