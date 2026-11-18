<?php
namespace Negative0 { trim('x'); }
namespace Negative1 { trim('x',' '); }
namespace Negative2 { trim('x','.x'); }
namespace Negative3 { trim('x','.1x'); }
namespace Negative4 { trim('x',$mask); }
namespace Negative5 { str_pad('abc',5); }
