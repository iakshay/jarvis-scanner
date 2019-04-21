import React, { Component } from 'react';
import './App.css';

class JobSnippet extends Component {
  constructor(props) {
    super(props);
  }

  render() {
      return <div>
      <div><strong> JobId </strong>{this.props.data.JobId}</div>
      <div><strong> Type </strong>{this.props.data.Type}</div>
      {this.props.data.Type === "PortScan" ? (
        <div>
        <div>{this.props.data.Data.Ip}</div>
        <div>{this.props.data.Data.PortRange.Start}-{this.props.data.Data.PortRange.End}</div>
        <div>{this.props.data.Data.Mode}</div>
      </div>) : (<div>{this.props.data.Data.IpBlock}</div>)}
      </div>
  }
}

class JobSubmit extends Component {
  render() {
    return (
      <div>
      JobDetailView
      </div>
    );
  }
}

class JobDetail extends Component {
  render() {
    return (
      <div>
      JobDetailView
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
    console.log('foo');
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
      <h1>ListView</h1>
      <ul>
          {this.state.data.map(job => (<JobSnippet data={job} />))}
      </ul>
      </div>
    );
  }
}

class App extends Component {
  render() {
    return (
      <div className="App">
        <JobList />
      </div>
    );
  }
}

export default App;
