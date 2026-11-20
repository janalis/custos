<?php
// presets: per; set: class_brace=keep, function_brace=same_line, closure_brace=next_line, control_brace=next_line, else_position=next_line, catch_position=keep, empty_body=keep, trim_block_blank_lines=true, class_open_blank_line=one, class_close_blank_line=one
class A {

    public function f() {
        if ($a)
        {
            $f = function ()
            {
                return 1;
            };
        }
        else
        {
        }
        try
        {
        } catch (E $e)
        {
        }
    }

}
$o = new class {

    public $x;

};
