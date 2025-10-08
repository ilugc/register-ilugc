package register_ilugc

  import (
          "log"
  )

  type Global struct {
          logger *log.Logger
  }
  var G *Global

  func init() {
          if G != nil {
                  return
          }
          G = &Global{}
          G.logger = log.Default()
          G.logger.SetFlags(log.Ldate |log.Ltime | log.Lmicroseconds | log.Lshortfile)
  }
