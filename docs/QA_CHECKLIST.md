# Sunpath v1.0 Manual QA Checklist

## Prerequisites
- [ ] Backend is running (`cd backend && go run ./cmd/sunpathd`)
- [ ] Frontend is running (`cd frontend && npm run dev`)
- [ ] App opens at http://localhost:5173

## Baseline (no pin)
- [ ] Page loads with empty map and "Click the map to drop a pin" message
- [ ] Map tiles render correctly
- [ ] Geocode search box is visible and functional
- [ ] No console errors on initial load

## Pin interaction
- [ ] Clicking on the map drops a red pin
- [ ] Coordinates appear in the side panel
- [ ] Pin can be dragged to a new location
- [ ] Map flies to the pin location on drop

## Geocode search
- [ ] Type "Paris" and see results appear
- [ ] Clicking a result moves the map and drops the pin
- [ ] Empty query shows no results
- [ ] Search box shows "Searching..." during request

## Observer height
- [ ] Slider changes height value (0-100m)
- [ ] Ground / Floor 3 / Floor 10 preset buttons work
- [ ] Active preset is visually highlighted
- [ ] Changing height re-fetches the horizon profile

## Horizon computation
- [ ] After dropping a pin, "Computing horizon profile..." appears
- [ ] Heatmap renders after computation completes
- [ ] Daily sun-hours strip appears below the heatmap
- [ ] Key dates (solstices, equinoxes, best/worst) are shown
- [ ] Plain-language summary is generated
- [ ] No errors in backend or frontend console

## Heatmap interaction
- [ ] Clicking on the heatmap highlights that day
- [ ] Selected day bar updates in the daily strip
- [ ] Selected day details (sun hours, first/last sun) update
- [ ] Arrow keys move selection left/right

## Key dates
- [ ] Clicking "Summer Solstice" loads that day
- [ ] Clicking "Best Day" loads the sunniest day
- [ ] All date buttons update the heatmap selection

## Low confidence
- [ ] If OSM data is thin, a yellow confidence banner appears
- [ ] Banner shows the percentage of estimated buildings
- [ ] Banner text is legible and informative

## URL sharing
- [ ] Dropping a pin updates the URL with lat/lng/h parameters
- [ ] Copying the URL and opening it in a new tab restores the analysis

## Error handling
- [ ] Offline mode shows "You are offline" banner
- [ ] If backend is down, error message with retry button appears
- [ ] Invalid pin (ocean area with no buildings) handles gracefully

## Accessibility
- [ ] Tab through all interactive elements works logically
- [ ] Heatmap has accessible ARIA label
- [ ] All buttons have visible focus states
- [ ] Color contrast meets WCAG AA for all text

## Method panel
- [ ] "About the method and its limits" link is visible
- [ ] Clicking it expands the explanation
- [ ] Explanation covers 2.5D model, height estimation, limitations

## Build
- [ ] `npm run build` completes without errors
- [ ] `go build ./cmd/sunpathd` completes without errors
- [ ] `go test ./...` passes all tests
