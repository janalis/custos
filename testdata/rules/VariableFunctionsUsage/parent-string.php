<?php
class Child extends Base {
    public function run($x) {
        call_user_func('parent::run', $x);
        call_user_func('PARENT::run', $x);
        call_user_func([$this, 'Parent::run'], $x);
        forward_static_call('parent::boot');
        return <weak_warning descr="Call it directly: 'Base::run($x)'.">call_user_func('Base::run', $x)</weak_warning>;
    }
}
