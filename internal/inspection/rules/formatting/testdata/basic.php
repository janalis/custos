<?php

namespace App;

function total($a,<info descr="Fix the spaces between tokens"></info>$b)<info descr="Move the function opening brace"></info>{<info descr="Fix the indentation">
  </info>if<info descr="Fix the spaces between tokens"></info>($a<info descr="Fix the spaces around the operator"></info>><info descr="Fix the spaces around the operator"></info>$b)<info descr="Move the opening brace"></info>{<info descr="Put the body on its own lines"></info>return $a;<info descr="Put the body on its own lines"></info>}
    return $b<info descr="Fix the spaces between tokens"> </info>;
}

class Cart
{
    private $items<info descr="Fix the spaces around the operator"></info>=<info descr="Fix the spaces around the operator"></info>[];
    public function add($item)<info descr="Move the function opening brace"></info>{
        $this->items[]<info descr="Fix the spaces around the operator"></info>=<info descr="Fix the spaces around the operator"></info>$item;
    }
}
