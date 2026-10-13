<?php
function route($verb) {
    <weak_warning descr="This switch has a single case; an 'if' is clearer.">switch</weak_warning> ($verb) {
        case 'GET':
            serve();
    }
    <weak_warning descr="This switch has a single case and a default; an 'if'/'else' is clearer.">switch</weak_warning> (strtoupper($verb)) {
        default:
            reject();
            break;
        case 'POST':
            store();
            break;
    }
    <weak_warning descr="This switch only has a default branch; keep just its body.">switch</weak_warning> ($verb):
        default:
            fallback();
    endswitch;
    <weak_warning descr="This switch has a single case; an 'if' is clearer.">SWITCH</weak_warning> ($verb) { case 1: }
}
