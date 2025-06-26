package audioStem

// finds an specific Frame from the Frames slice
func GetFrame(frames []*AudioStemThread_Frame, filename string) *AudioStemThread_Frame {
	for _, frame := range frames {
		if frame.Filename == filename {
			return frame
		}
	}
	return nil
}
