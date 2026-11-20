<?php
// presets: per symfony laravel drupal psr1
/**
 * File docblock.
 *   indented badly
 */

# hash comment
namespace App;

/* block */ use A; // trailing

class Foo // after name
{
    /**
     * Doc.
     */
    public $a; # hash after

    // before method
    public function f($x /* inline */, $y) // trailing
    {
        $a = [ // open
            1, // one
            /* two */ 2,
        ];
        /*
          free form
            text
        */
        if ($x) { // c
            return; // done
        } // after
        // before else
        else {
        }
        foo(
            // leading comment
            $a,
        );
        return $x
            // between
            + $y;
    }
    // trailing in class
}
// end of file
