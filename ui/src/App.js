import React, { Component } from 'react';
import { BrowserRouter as Router, Route, Link } from "react-router-dom";

class JobSnippet extends Component {
  render() {
      return (<div className="job-snippet row mb-10 p-3">
      <div className="col-10">
      <div><strong>#</strong>{this.props.data.JobId} / <strong>Type: </strong>{this.props.data.Type}</div>
      {this.props.data.Type === "PortScan" ? 
        (<div>Mode: {this.props.data.Data.Mode} / Ip: {this.props.data.Data.Ip} / Ports: {this.props.data.Data.PortRange.Start}-{this.props.data.Data.PortRange.End}</div>)
        : (<div>Ip: {this.props.data.Data.IpBlock}</div>)}
        </div>
      <div className="col-2">
        <Link className="btn btn-primary mr-1" to={`/view/${this.props.data.JobId}`}>detail</Link>
        <button type="button" className="btn btn-danger mr-1">delete</button>
      </div>
      </div>)
  }
}

class TaskView extends Component {
  render() {
      return (<div>{JSON.stringify(this.props.data)}</div>)
  }
}


class JobSubmit extends Component {
  constructor(props) {
    super(props);

    this.state = {
      jobType: "PortAlive"
    }
  }
  handleChange = (e) => {
        console.log(e.target.name, e.target.value);
        this.setState({[e.target.name] : e.target.value});
    };
  render() {
    return (
      <div className="p-3">
      <h3>Submit job </h3>
      <form>
          <div className="row form-inline mb-3">
           <div className="form-group col-2 mr-3">
           <label for="jobType" className="col-sm-2">Type</label>
            <div className="col-sm-4">
            <select style={{width:'120px'}} className="custom-select"
                    defaultValue="IsAlive"
                    name="jobType" onChange={this.handleChange}>
                <option value="IsAlive">IsAlive</option>
                <option value="PortScan">PortScan</option>
            </select>
            </div>
            </div>
          <div className="form-group col-6 mr-2">
          <label for="ipBlock" className="col-sm-2">IpBlock</label>
          <div className="col-sm-4">
          <input style={{width:'215px'}}className="form-control"
                         type="text"
                         name="ipBlock"
                         placeholder="192.168.0.1"
                         onChange={this.handleChange} />
           </div>
           </div>
           </div>
          {this.state.jobType === "PortScan" ?
          (<div className="row form-inline mb-3">
            <div className="form-group col-2 mr-3">
            <label for="mode" className="col-sm-2">Mode</label>
             <div className="col-sm-4">
            <select style={{width:'120px'}} className="custom-select"
                    defaultValue="NormalScan"
                    name="mode" onChange={this.handleChange}>
                <option value="NormalScan">Normal</option>
                <option value="SynScan">SYN</option>
                <option value="FinScan">FIN</option>
            </select>
            </div>
            </div>
            <div className="form-group col-6 mr-3">
            <label for="portMin" className="col-sm-2">Ports</label>
            <div className="col-sm-6">
            <input style={{width:'100px'}} className="form-control "
                           type="number"
                           name="portMin"
                           defaultValue="0"
                           onChange={this.handleChange} />
                           <span> - </span>
            <input style={{width:'100px'}} className="form-control"
                           type="number"
                           name="portMax"
                           defaultValue="65535"
                           onChange={this.handleChange} /></div></div></div>) : '' }

          <button type="button" className="btn btn-primary">submit</button>
      </form>
      </div>
    );
  }
}

class JobDetail extends Component {
  constructor(props) {
    super(props);
    this.state = {data: {}};
  }

  componentDidMount() {
    console.log('foo');
    // fetch jobs detail here
    this.setState({'data':
    {
      JobId: 1,
      Data: [
      {
        WorkerId: 1,
        WorkerName: "Worker1",
        Data: {
          Ip: "192.168.2.1",
          Status: 0
        }
      },
      {
        WorkerId: 2,
        WorkerName: "Worker2",
        Data: {
          Ip: "192.168.2.1",
          Status: 0,
        },
      },
      {
        WorkerId: 3,
        WorkerName: "Worker3",
        Data: {
          Ip: "192.168.2.1",
          Status: 0,
        },
      },
      {
        WorkerId: 4,
        WorkerName: "Worker4",
        Data: {
          Ip: "192.168.2.1",
          Status: 0,
        }
      }]
    }});
  }
  
  render() {
    console.log(this.state);
    return (
      <div>
      JobDetailView {this.props.match.params.id}
      <div>
          {this.state.data.Data && this.state.data.Data.map(task => (<TaskView data={task} />))}
      </div>
      </div>
    );
  }
}

class JobList extends Component {
  constructor(props) {
    super(props);
    this.state = {data: []};
  }
  
  componentDidMount() {
    // fetch jobs here
    this.setState({'data':
    [{
      JobId: 1,
      Type: "PortScan",
      Data: {
          Ip: "192.168.0.1",
          PortRange: {Start: 1000, End: 2000},
          Mode: "SynScan"
      }
    },
    {
      JobId: 2,
      Type: "IsAlive",
      Data: {
        IpBlock: "192.168.0.1",
      }
    },
    {
      JobId: 3,
      Type: "IsAlive",
      Data: {
        IpBlock: "192.168.0.1/24",
      }
    }]});
  }
  
  render() {
    return (
      <div>
          {this.state.data.map(job => (<JobSnippet data={job} />))}
      </div>
    );
  }
}

function Home() {
  return (<div><JobSubmit /><JobList /></div>)
}

function About() {
  return <h3>About</h3>;
}

class App extends Component {
  render() {
    return (
      <div className="App">
      <Router>
        <nav className="navbar navbar-expand-lg navbar-light bg-light">
          <span className="navbar-brand mb-0 h1">Jarvis Scanner</span>
            <ul className="navbar-nav mr-auto">
             <li className="nav-item active">
                <Link className="nav-link active" to='/'>Home</Link>
             </li>
            <li className="nav-item">
                <Link className="nav-link" to='/about'>About</Link>
            </li>
                
            </ul>
        </nav>
        <div className="container">
          <Route path="/" exact component={Home} />
          <Route path="/about/" component={About} />
          <Route path='/view/:id' component={JobDetail} />
        </div>
      </Router>
      </div>
    );
  }
}

export default App;
