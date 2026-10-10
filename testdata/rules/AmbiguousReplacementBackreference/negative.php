<?php
namespace Negative0 { preg_replace('/(r)/','${1}1','r'); }
namespace Negative1 { preg_replace('/x/','$11','x'); }
namespace Negative2 { preg_replace($pattern,'$11','x'); }
namespace Negative3 { preg_replace('/(r)/','$1','r'); }
namespace Negative4 { preg_replace('/(?|x)/','$11','x'); }
namespace Negative5 { preg_replace('/(r)/',$replacement,'r'); }
namespace Negative6 { trim($s); }
namespace Negative7 { preg_replace('/(r)/','\\$11','r'); }

namespace Audit0 { preg_replace('/\\Q(a)\\E(b)/','$21','(a)b'); }
