package main



func modules() ([]Module, []Module){
  left_gen := [](func() Module){NewHyprWorkspaces, NewHyprTitle}
  right_gen := [](func() Module){NewClock}


  var left, right []Module

  for _, genmod := range left_gen {
    left = append(left, genmod())
  }
  for _, genmod := range right_gen {
    right = append(right, genmod())
  }

  return left, right
}
