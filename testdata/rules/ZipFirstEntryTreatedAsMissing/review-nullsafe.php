<?php
class ReviewHolder { public ?ZipArchive $zip = null; public ZipArchive $open; }
function reviewNullsafe(?ZipArchive $z, ?ReviewHolder $h) {
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">!$z?->locateName('a')</warning>) {
        echo "missing";
    }
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">$z?->locateName('a')</warning>) {
        echo "found";
    }
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">$z?->locateName('a') == false</warning>) {
        echo "missing";
    }
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">$h?->open->locateName('a')</warning>) {
        echo "found";
    }
}
