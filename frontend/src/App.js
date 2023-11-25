import './App.css';
import { BrowserRouter as Router, Route, Switch } from 'react-router-dom';
import SignIn from './pages/SignIn';
import SignUp from './pages/SignUp';

function App() {
  return (
    <div className="App">
      <Router>
		<Switch>
			<Route path="/signin" component={SignIn} />
			<Route path="/signup" component={SignUp} />
		</Switch>
		</Router>
    </div>
  );
}

export default App;
